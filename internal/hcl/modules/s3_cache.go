package modules

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mholt/archiver/v3"
)

type S3Cache struct {
	s3Client   *s3.S3
	bucketName string
	prefix     string
}

// NewS3Cache creates a new S3Cache instance
func NewS3Cache(region, bucketName, prefix string) (*S3Cache, error) {
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
		s3Client:   s3.New(sess),
		bucketName: bucketName,
		prefix:     prefix,
	}, nil
}

func (cache *S3Cache) applyPrefix(key string) string {
	// URL encode the key first since the module address contains /'s
	// and these create folders in S3
	encodedKey := url.QueryEscape(key)

	if cache.prefix != "" {
		return fmt.Sprintf("%s/%s", cache.prefix, encodedKey)
	}
	return encodedKey
}

// Exists checks if the key exists in the S3 bucket
func (cache *S3Cache) Exists(key string) (bool, error) {
	prefixedKey := cache.applyPrefix(key)
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
func (cache *S3Cache) Get(key, destPath string) error {
	prefixedKey := cache.applyPrefix(key)

	// Download from S3
	result, err := cache.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(cache.bucketName),
		Key:    aws.String(prefixedKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	// Create a temporary file for the downloaded archive
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("s3cache-%s.tar.gz", uuid.New().String()))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the S3 object to the temporary file
	if _, err := io.Copy(tmpFile, result.Body); err != nil {
		return fmt.Errorf("failed to save downloaded file: %w", err)
	}

	// Extract using archiver
	tgz := archiver.NewTarGz()
	tgz.OverwriteExisting = true
	if err := tgz.Unarchive(tmpFile.Name(), destPath); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// Put uploads the srcPath to the S3 bucket with the key
func (cache *S3Cache) Put(key, srcPath string, ttl time.Duration) error {
	prefixedKey := cache.applyPrefix(key)

	// Generate a temporary file path without creating the file
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("s3cache-%s.tar.gz", uuid.New().String()))
	defer os.Remove(tmpPath)

	tgz := archiver.NewTarGz()

	// Get the contents of the source directory and create a list of paths to archive
	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	sources := make([]string, 0, len(entries))
	for _, entry := range entries {
		sources = append(sources, filepath.Join(srcPath, entry.Name()))
	}

	if err := tgz.Archive(sources, tmpPath); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	// Upload the archive to S3
	file, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Calculate expiration time and set it as metadata
	expirationTime := time.Now().Add(ttl).Format(time.RFC3339)
	metadata := map[string]*string{
		"x-amz-meta-expires-at": aws.String(expirationTime),
	}

	_, err = cache.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:   aws.String(cache.bucketName),
		Key:      aws.String(prefixedKey),
		Body:     file,
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}
