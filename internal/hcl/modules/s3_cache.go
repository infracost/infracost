package modules

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/aws/amazon-s3-encryption-client-go/v3/client"
	"github.com/aws/amazon-s3-encryption-client-go/v3/materials"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go/transport/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mholt/archiver/v3"
)

type S3Cache struct {
	s3Client   *s3.Client
	encClient  *client.S3EncryptionClientV3
	bucketName string
	prefix     string
}

type S3CacheConfig struct {
	Region        string
	BucketName    string
	Prefix        string
	KMSKeyID      string
	UseEncryption bool
}

// NewS3Cache creates a new S3Cache instance
func NewS3Cache(c S3CacheConfig) (*S3Cache, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(c.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	cache := &S3Cache{
		s3Client:   s3Client,
		bucketName: c.BucketName,
		prefix:     c.Prefix,
	}

	if c.UseEncryption {
		kmsClient := kms.NewFromConfig(cfg)
		cmm, err := materials.NewCryptographicMaterialsManager(materials.NewKmsKeyring(kmsClient, c.KMSKeyID))
		if err != nil {
			return nil, fmt.Errorf("failed to create KMS keyring: %w", err)
		}

		encClient, err := client.New(s3Client, cmm)
		if err != nil {
			return nil, fmt.Errorf("failed to create encryption client: %w", err)
		}

		cache.encClient = encClient
	}

	return cache, nil
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
	_, err := cache.s3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(cache.bucketName),
		Key:    aws.String(prefixedKey),
	})
	if err != nil {
		var respErr *http.ResponseError
		if errors.As(err, &respErr) {
			if respErr.HTTPStatusCode() == 404 || respErr.HTTPStatusCode() == 403 {
				return false, nil
			}
		}

		var notFound *types.NotFound
		var noSuchBucket *types.NoSuchBucket
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &notFound) || errors.As(err, &noSuchBucket) || errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Get downloads the key from the S3 bucket to the destPath
func (cache *S3Cache) Get(key, destPath string) error {
	prefixedKey := cache.applyPrefix(key)
	ctx := context.Background()

	var result *s3.GetObjectOutput
	var err error

	if cache.encClient != nil {
		result, err = cache.encClient.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(cache.bucketName),
			Key:    aws.String(prefixedKey),
		})
	} else {
		result, err = cache.s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(cache.bucketName),
			Key:    aws.String(prefixedKey),
		})
	}

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
	ctx := context.Background()

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
	input := &s3.PutObjectInput{
		Bucket: aws.String(cache.bucketName),
		Key:    aws.String(prefixedKey),
		Body:   file,
		Metadata: map[string]string{
			"x-amz-meta-expires-at": expirationTime,
		},
	}

	if cache.encClient != nil {
		_, err = cache.encClient.PutObject(ctx, input)
	} else {
		_, err = cache.s3Client.PutObject(ctx, input)
	}

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}
