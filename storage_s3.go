package burrow

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

var _ Storage = (*S3Storage)(nil)

// S3Storage implements the Storage interface for AWS S3
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates a new S3Storage instance
func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{
		client: client,
		bucket: bucket,
	}
}

// PutObject implements Storage.PutObject
func (s *S3Storage) PutObject(ctx context.Context, key string, reader io.Reader, contentType string, contentLength int64, metadata map[string]string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	}
	// Only set ContentLength if it's not -1
	if contentLength >= 0 {
		input.ContentLength = aws.Int64(contentLength)
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return convertS3Error(err)
	}
	return nil
}

// GetObject implements Storage.GetObject
func (s *S3Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, nil, convertS3Error(err)
	}
	info := &ObjectInfo{
		ContentType:     aws.ToString(result.ContentType),
		ContentLength:   aws.ToInt64(result.ContentLength),
		ContentEncoding: aws.ToString(result.ContentEncoding),
		ContentLanguage: aws.ToString(result.ContentLanguage),
		LastModified:    aws.ToTime(result.LastModified),
		ModTime:         aws.ToTime(result.LastModified),
		ETag:            aws.ToString(result.ETag),
		ChecksumSHA256:  aws.ToString(result.ChecksumSHA256),
		Metadata:        result.Metadata,
		Exists:          true,
	}
	return result.Body, info, nil
}

// HeadObject implements Storage.HeadObject
func (s *S3Storage) HeadObject(ctx context.Context, key string) (*ObjectInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, convertS3Error(err)
	}
	return &ObjectInfo{
		ContentType:     aws.ToString(result.ContentType),
		ContentLength:   aws.ToInt64(result.ContentLength),
		ContentEncoding: aws.ToString(result.ContentEncoding),
		ContentLanguage: aws.ToString(result.ContentLanguage),
		LastModified:    aws.ToTime(result.LastModified),
		ModTime:         aws.ToTime(result.LastModified),
		ETag:            aws.ToString(result.ETag),
		ChecksumSHA256:  aws.ToString(result.ChecksumSHA256),
		Metadata:        result.Metadata,
		Exists:          true,
	}, nil
}

// SignURL implements Storage.SignURL
func (s *S3Storage) SignURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	result, err := presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return "", convertS3Error(err)
	}
	return result.URL, nil
}

// convertS3Error converts AWS S3 errors to StorageError
func convertS3Error(err error) error {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch {
		case ae.ErrorCode() == "NoSuchKey" || ae.ErrorCode() == "NotFound":
			return &StorageError{Code: StorageErrNotFound, Message: ae.Error()}
		case ae.ErrorCode() == "AccessDenied":
			return &StorageError{Code: StorageErrAccess, Message: ae.Error()}
		default:
			return &StorageError{Code: StorageErrInternal, Message: ae.Error()}
		}
	}
	return &StorageError{Code: StorageErrInternal, Message: err.Error()}
}
