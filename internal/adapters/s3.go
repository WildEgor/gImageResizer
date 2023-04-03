package adapters

import (
	"context"
	"errors"
	"io"

	"github.com/WildEgor/gImageResizer/internal/configs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type S3Obj struct {
	BucketName  string
	Key         string
	Size        int64
	ContentType string
	Reader      *io.Reader
	Bytes       []byte
}

type IS3Adapter interface {
	PutObj(ctx context.Context, obj *S3Obj) error
}

type S3Adapter struct {
	client *minio.Client
	config *configs.S3Config
}

func NewS3Adapter(
	config *configs.S3Config,
) *S3Adapter {

	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = initBucket(context.Background(), config.Bucket, config.Region, minioClient)
	if err != nil {
		log.Fatalln(err)
	}

	return &S3Adapter{
		client: minioClient,
		config: config,
	}
}

func (m *S3Adapter) PutObj(ctx context.Context, obj *S3Obj) error {
	data := S3Obj(*obj)

	if obj.BucketName == "" {
		data.BucketName = m.config.Bucket
	}

	if obj.ContentType == "" {
		return errors.New("[S3Adapter] PutObj empty content-type not allowed")
	}

	info, err := m.client.PutObject(ctx, data.BucketName, data.Key, *data.Reader, data.Size, minio.PutObjectOptions{
		ContentType: data.ContentType,
	})
	if err != nil {
		return errors.New("[S3Adapter] PutObj failed to put")
	}

	log.Printf("[S3Adapter] Successfully uploaded %s of size %d\n", data.Key, info.Size)

	return nil
}

func initBucket(ctx context.Context, bucketName string, region string, client *minio.Client) error {
	err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region})
	if err != nil {

		exists, errBucketExists := client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("[S3Adapter] We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}

	} else {
		log.Printf("[S3Adapter] Successfully created %s\n", bucketName)
	}

	return nil
}
