package modules

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeObject struct {
	body   []byte
	header http.Header
}

// fakeS3 is a minimal path-style S3 server supporting HEAD, GET and PUT.
type fakeS3 struct {
	mu        sync.Mutex
	objects   map[string]fakeObject
	forbidden map[string]bool
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := r.URL.EscapedPath()
	if f.forbidden[key] {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		h := http.Header{}
		for k, v := range r.Header {
			if strings.HasPrefix(strings.ToLower(k), "x-amz-meta-") {
				h[k] = v
			}
		}
		f.objects[key] = fakeObject{body: body, header: h}
		w.WriteHeader(http.StatusOK)
	case http.MethodHead, http.MethodGet:
		obj, ok := f.objects[key]
		if !ok {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			if r.Method == http.MethodGet {
				_, _ = w.Write([]byte(`<Error><Code>NoSuchKey</Code></Error>`))
			}
			return
		}
		for k, v := range obj.header {
			w.Header()[k] = v
		}
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write(obj.body)
		}
	}
}

func newTestS3Cache(t *testing.T, f *fakeS3) *S3Cache {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)

	client := s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider("id", "secret", ""),
	})

	return &S3Cache{
		s3Client:     client,
		bucketName:   "bucket",
		prefix:       "private",
		publicPrefix: "publicModules",
		cachePrivate: true,
	}
}

func TestS3CacheRoundTrip(t *testing.T) {
	f := &fakeS3{objects: map[string]fakeObject{}, forbidden: map[string]bool{}}
	cache := newTestS3Cache(t, f)

	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "main.tf"), []byte("resource \"x\" \"y\" {}"), 0600))

	exists, err := cache.Exists("git::example.com/mod", false)
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, cache.Put("git::example.com/mod", src, time.Hour, false))

	exists, err = cache.Exists("git::example.com/mod", false)
	require.NoError(t, err)
	assert.True(t, exists)

	dest := t.TempDir()
	require.NoError(t, cache.Get("git::example.com/mod", dest, false))
	b, err := os.ReadFile(filepath.Join(dest, "main.tf"))
	require.NoError(t, err)
	assert.Equal(t, "resource \"x\" \"y\" {}", string(b))
}

func TestS3CacheExpired(t *testing.T) {
	f := &fakeS3{objects: map[string]fakeObject{}, forbidden: map[string]bool{}}
	cache := newTestS3Cache(t, f)

	require.NoError(t, cache.Put("mod", t.TempDir(), -time.Minute, true))

	exists, err := cache.Exists("mod", true)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestS3CacheForbidden(t *testing.T) {
	f := &fakeS3{objects: map[string]fakeObject{}, forbidden: map[string]bool{"/bucket/publicModules/mod": true}}
	cache := newTestS3Cache(t, f)

	exists, err := cache.Exists("mod", true)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestS3CachePrivateDisabled(t *testing.T) {
	f := &fakeS3{objects: map[string]fakeObject{}, forbidden: map[string]bool{}}
	cache := newTestS3Cache(t, f)
	cache.cachePrivate = false

	require.NoError(t, cache.Put("mod", t.TempDir(), time.Hour, false))
	assert.Empty(t, f.objects)
}
