package aws

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/olad5/go-cloud-backup-system/config"
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
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	return result.Location, nil
}

func (a *AwsFileStore) GetOne(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (a *AwsFileStore) Ping(ctx context.Context) error {
	return nil
}
