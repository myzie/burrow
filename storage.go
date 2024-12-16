package burrow

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Storage represents an abstract object storage interface
type Storage interface {
	// PutObject writes data to the storage with the given key
	PutObject(ctx context.Context, key string, reader io.Reader, contentType string, contentLength int64, metadata map[string]string) error

	// GetObject retrieves an object from storage
	GetObject(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error)

	// HeadObject checks if an object exists and returns its metadata
	HeadObject(ctx context.Context, key string) (*ObjectInfo, error)

	// SignURL generates a pre-signed URL for the object with the given key
	SignURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

// ObjectInfo contains metadata about a stored object
type ObjectInfo struct {
	ContentType     string
	ContentLength   int64
	ContentEncoding string
	ContentLanguage string
	ContentLocation string
	ChecksumSHA256  string
	LastModified    time.Time
	ModTime         time.Time
	Exists          bool
	ETag            string
	Metadata        map[string]string
}

// StorageError represents storage-specific errors
type StorageError struct {
	Code    string
	Message string
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common storage error codes
const (
	StorageErrNotFound = "NotFound"
	StorageErrAccess   = "AccessDenied"
	StorageErrInternal = "InternalError"
)
