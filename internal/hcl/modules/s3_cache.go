package modules

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mholt/archives"
)

type S3Cache struct {
	s3Client     *s3.S3
	bucketName   string
	prefix       string
	publicPrefix string
	cachePrivate bool
}

// NewS3Cache creates a new S3Cache instance
func NewS3Cache(region, bucketName, prefix string, cachePrivate bool) (*S3Cache, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(region),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Cache{
		s3Client:     s3.New(sess),
		bucketName:   bucketName,
		prefix:       prefix,
		publicPrefix: "publicModules",
		cachePrivate: cachePrivate,
	}, nil
}

func (cache *S3Cache) applyPrefix(key string, public bool) string {
	// URL encode the key first since the module address contains /'s
	// and these create folders in S3
	encodedKey := url.QueryEscape(key)

	// if its public, put it in the public prefix
	if public {
		return fmt.Sprintf("%s/%s", cache.publicPrefix, encodedKey)
	}

	if cache.prefix != "" {
		return fmt.Sprintf("%s/%s", cache.prefix, encodedKey)
	}
	return encodedKey
}

// Exists checks if the key exists in the S3 bucket
func (cache *S3Cache) Exists(key string, public bool) (bool, error) {
	if !public && !cache.cachePrivate {
		return false, nil
	}

	prefixedKey := cache.applyPrefix(key, public)
	headObj, err := cache.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(cache.bucketName),
		Key:    aws.String(prefixedKey),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey, "NotFound", "Forbidden":
				return false, nil
			}
		}
		return false, err
	}

	// Check expiration based on x-amz-meta-expires-at
	if expiresAtStr, ok := headObj.Metadata["x-amz-meta-expires-at"]; ok {
		expirationTime, err := time.Parse(time.RFC3339, *expiresAtStr)
		if err != nil {
			return false, fmt.Errorf("failed to parse expiration metadata: %w", err)
		}

		if time.Now().After(expirationTime) {
			return false, nil // Object is expired
		}
	}

	return true, nil
}

// Get downloads the key from the S3 bucket to the destPath
func (cache *S3Cache) Get(key, destPath string, public bool) error {
	if !public && !cache.cachePrivate {
		logger.Debug("Cache is disabled for private modules")
		return nil
	}

	prefixedKey := cache.applyPrefix(key, public)

	// Download from S3
	result, err := cache.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(cache.bucketName),
		Key:    aws.String(prefixedKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer func() { _ = result.Body.Close() }()

	// Ensure the destination directory exists
	if err := os.MkdirAll(destPath, 0700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	ctx := context.Background()

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Extraction:  archives.Tar{},
	}

	// Extract the archive
	err = format.Extract(ctx, result.Body, func(_ context.Context, f archives.FileInfo) error {

		// no symlinks thank you
		if f.LinkTarget != "" {
			return nil
		}

		// ensure the archive isn't maliciously pathing outside of the destination directory
		name := filepath.Clean(f.NameInArchive)

		// skip git files, we don't need them
		if strings.HasPrefix(name, ".git/") {
			return nil
		}

		// Determine where to create this file on disk
		targetPath := filepath.Join(destPath, name)
		// For directories, just create them
		if f.IsDir() {
			return os.MkdirAll(targetPath, 0700)
		}

		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0700); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Create the file
		// #nosec G304
		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer func() { _ = outFile.Close() }()

		// Copy the contents
		reader, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}
		defer func() { _ = reader.Close() }()

		_, err = io.Copy(outFile, reader)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// Put uploads the srcPath to the S3 bucket with the key
func (cache *S3Cache) Put(key, srcPath string, ttl time.Duration, public bool) error {
	if !public && !cache.cachePrivate {
		logger.Debug("Cache is disabled for private modules")
		return nil
	}

	prefixedKey := cache.applyPrefix(key, public)

	// Create a context for the archiving
	ctx := context.Background()

	// Get the contents of the source directory
	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Create a map of disk files to archive paths
	fileMap := make(map[string]string)
	for _, entry := range entries {
		diskPath := filepath.Join(srcPath, entry.Name())
		archivePath := entry.Name()
		fileMap[diskPath] = archivePath
	}

	// Get a list of files to archive
	files, err := archives.FilesFromDisk(ctx, nil, fileMap)
	if err != nil {
		return fmt.Errorf("failed to get files for archiving: %w", err)
	}

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}

	// Generate a temporary file path
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("s3cache-%s.tar.gz", uuid.New().String()))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	// Create the archive
	if err := format.Archive(ctx, tmpFile, files); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	// Calculate expiration time and set it as metadata
	expirationTime := time.Now().Add(ttl).Format(time.RFC3339)
	metadata := map[string]*string{
		"x-amz-meta-expires-at": aws.String(expirationTime),
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to start of file: %w", err)
	}

	_, err = cache.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:   aws.String(cache.bucketName),
		Key:      aws.String(prefixedKey),
		Body:     tmpFile,
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}
