package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/olad5/file-fort/config"
)

type AwsFileStore struct {
	Client  *s3.S3
	Bucket  *string
	session *session.Session
}

func NewAwsFileStore(ctx context.Context, configurations *config.Configurations) (*AwsFileStore, error) {
	region := configurations.AwsRegion

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(configurations.AwsAccessKey, configurations.AwsSecretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(configurations.AwsEndpoint),
	})
	if err != nil {
		return &AwsFileStore{}, fmt.Errorf("error creating new s3 session: %w", err)
	}

	svc := s3.New(sess)

	bucket := aws.String(configurations.AwsS3Bucket)
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: bucket,
	})

	if err != nil {
		return &AwsFileStore{}, fmt.Errorf("Unable to create S3 bucket %q, %v", bucket, err)
	}

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: bucket,
	})

	if err != nil {
		return &AwsFileStore{}, fmt.Errorf("Unable to create S3 bucket after Waiting %q, %v", bucket, err)
	}

	return &AwsFileStore{
		Client:  svc,
		Bucket:  bucket,
		session: sess,
	}, nil
}

func (a *AwsFileStore) SaveToFileStore(ctx context.Context, filename string, file io.Reader) (string, error) {
	uploader := s3manager.NewUploader(a.session)
	bucket := a.Bucket
	key := filename
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	return key, nil
}

func (a *AwsFileStore) GetDownloadUrl(ctx context.Context, key string) (string, error) {
	downloadUrl, err := a.generatePreSignedUrl(ctx, key)
	if err != nil {
		return "", fmt.Errorf("error getting download url", err)
	}
	return downloadUrl, nil
}

func (a *AwsFileStore) generatePreSignedUrl(ctx context.Context, key string) (string, error) {
	req, _ := a.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: a.Bucket,
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", fmt.Errorf("Error generating presigned url: %w", err)
	}

	return urlStr, nil
}

func (a *AwsFileStore) DeleteFile(ctx context.Context, key string) error {
	_, err := a.Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: a.Bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error deleting file from file store", err)
	}
	return nil
}

func (a *AwsFileStore) Ping(ctx context.Context) error {
	return nil
}
