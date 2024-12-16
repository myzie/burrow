package burrow

import (
	"context"
	"errors"
	"io"
	"time"

	"bytes"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	// If content length is known, use simple PutObject
	if contentLength >= 0 {
		return s.simplePutObject(ctx, key, reader, contentType, contentLength, metadata)
	}
	// For unknown length, use multipart upload
	return s.multipartPutObject(ctx, key, reader, contentType, metadata)
}

func (s *S3Storage) simplePutObject(ctx context.Context, key string, reader io.Reader, contentType string, contentLength int64, metadata map[string]string) error {
	// Current implementation
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		Metadata:      metadata,
		ContentLength: aws.Int64(contentLength),
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return convertS3Error(err)
	}
	return nil
}

func (s *S3Storage) multipartPutObject(ctx context.Context, key string, reader io.Reader, contentType string, metadata map[string]string) error {
	// Initialize multipart upload
	createInput := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	}
	createResult, err := s.client.CreateMultipartUpload(ctx, createInput)
	if err != nil {
		return convertS3Error(err)
	}
	// If the upload fails, make sure to abort it
	defer func() {
		if err != nil {
			abortInput := &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s.bucket),
				Key:      aws.String(key),
				UploadId: createResult.UploadId,
			}
			_, _ = s.client.AbortMultipartUpload(ctx, abortInput)
		}
	}()
	// Upload parts
	var completedParts []types.CompletedPart
	partNumber := int32(1)
	buffer := make([]byte, 5*1024*1024) // 5MB buffer size (minimum for S3 multipart)
	for {
		// Read a chunk
		n, readErr := io.ReadFull(reader, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			err = readErr
			return convertS3Error(err)
		}
		// If we read some data, upload it as a part
		if n > 0 {
			uploadInput := &s3.UploadPartInput{
				Bucket:     aws.String(s.bucket),
				Key:        aws.String(key),
				PartNumber: aws.Int32(partNumber),
				UploadId:   createResult.UploadId,
				Body:       bytes.NewReader(buffer[:n]),
			}
			uploadResult, uploadErr := s.client.UploadPart(ctx, uploadInput)
			if uploadErr != nil {
				err = uploadErr
				return convertS3Error(err)
			}
			completedParts = append(completedParts, types.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int32(partNumber),
			})
			partNumber++
		}
		// If we've reached EOF, break
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
	}
	// Complete multipart upload
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: createResult.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	_, err = s.client.CompleteMultipartUpload(ctx, completeInput)
	if err != nil {
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
