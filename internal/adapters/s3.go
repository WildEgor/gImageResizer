package adapters

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/WildEgor/gImageResizer/internal/configs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

type S3Obj struct {
	Bucket        string
	Key           string
	ContentLength int64
	ContentType   string
	Body          *io.ReadSeeker
	Bytes         []byte
	PartNumber    int64
}

type IS3Adapter interface {
	PutObj(ctx context.Context, obj *S3Obj) error
	SessionUpload(ctx context.Context, obj *S3Obj) (*string, error)
	GetPresign(ctx context.Context, obj *S3Obj) (*string, error)
}

type S3Adapter struct {
	client *s3.S3
	config *configs.S3Config
}

func NewS3Adapter(
	config *configs.S3Config,
) *S3Adapter {

	creds := credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, "")
	_, err := creds.Get()
	if err != nil {
		log.Error(err)
		log.Fatal("[S3Adapter] Bad creds")
	}

	cfg := aws.NewConfig().WithRegion(config.Region).WithCredentials(creds)
	ss, err := session.NewSession(cfg)

	if err != nil {
		log.Fatal("[S3Adapter] Failed init session")
	}

	client := s3.New(ss, cfg)

	return &S3Adapter{
		client: client,
		config: config,
	}
}

func (m *S3Adapter) PutObj(ctx context.Context, obj *S3Obj) error {
	data := S3Obj(*obj)

	if obj.Bucket == "" {
		data.Bucket = m.config.Bucket
	}

	if obj.ContentType == "" {
		return errors.New("[S3Adapter] PutObj empty content-type not allowed")
	}

	_, err := m.client.PutObject(&s3.PutObjectInput{
		Body:          *data.Body,
		Key:           &data.Key,
		ContentType:   &data.ContentType,
		ContentLength: &data.ContentLength,
		Bucket:        &data.Bucket,
	})

	if err != nil {
		return errors.New("[S3Adapter] PutObj failed to put")
	}

	return nil
}

func (m *S3Adapter) GetPresign(
	ctx context.Context,
	obj *S3Obj,
) (*string, error) {
	data := S3Obj(*obj)

	if obj.Bucket == "" {
		data.Bucket = m.config.Bucket
	}

	req, _ := m.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &data.Bucket,
		Key:    &data.Key,
	})

	link, err := req.Presign(time.Duration(168) * time.Hour)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (m *S3Adapter) SessionUpload(
	ctx context.Context,
	obj *S3Obj,
) (*string, error) {
	data := S3Obj(*obj)

	if obj.Bucket == "" {
		data.Bucket = m.config.Bucket
	}

	if obj.ContentType == "" {
		return nil, errors.New("[S3Adapter] Empty content-type not allowed")
	}

	resp, err := m.client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      &data.Bucket,
		Key:         &data.Key,
		ContentType: &data.ContentType,
	})
	if err != nil {
		log.Errorf("[S3Adapter] Failed %v", err.Error())
		return nil, err
	}
	log.Debug("[S3Adapter] Created multipart upload request...")

	var curr, partLength int64
	var remaining = data.ContentLength
	var completedParts []*s3.CompletedPart
	partNumber := 1
	maxPartSize := int64(5 * 1024 * 1024)

	log.Debug(remaining)

	for curr = 0; remaining != 0; curr += partLength {
		if remaining < maxPartSize {
			partLength = remaining
		} else {
			partLength = maxPartSize
		}
		log.Debug(partLength)
		// Upload binaries part
		completedPart, err := m.uploadPart(resp, data.Bytes[curr:curr+partLength], partNumber)

		// If upload this part fail
		// Make an abort upload error and exit
		if err != nil {
			log.Errorf("[S3Adapter] Failed %v", err.Error())
			err := m.abortMultipartUpload(resp)
			if err != nil {
				log.Errorf("[S3Adapter] Failed %v", err.Error())
			}
			return nil, err
		}
		// else append completed part to a whole
		remaining -= partLength
		partNumber++
		completedParts = append(completedParts, completedPart)
	}

	completeResponse, err := m.completeMultipartUpload(resp, completedParts)
	if err != nil {
		log.Errorf("[S3Adapter] Failed %v", err.Error())
		return nil, err
	}

	log.Debug("[S3Adapter] Successfully uploaded file: %s\n", completeResponse.String())

	return completeResponse.Location, nil
}

func (m *S3Adapter) completeMultipartUpload(
	resp *s3.CreateMultipartUploadOutput,
	completedParts []*s3.CompletedPart,
) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	return m.client.CompleteMultipartUpload(completeInput)
}

func (m *S3Adapter) abortMultipartUpload(resp *s3.CreateMultipartUploadOutput) error {
	log.Debug("[S3Adapter] Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := m.client.AbortMultipartUpload(abortInput)
	return err
}

func (m *S3Adapter) uploadPart(
	resp *s3.CreateMultipartUploadOutput,
	fileBytes []byte,
	partNumber int,
) (*s3.CompletedPart, error) {
	tryNum := 1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	for tryNum <= 3 {
		uploadResult, err := m.client.UploadPart(partInput)
		if err != nil {
			if tryNum == 3 {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			log.Debugf("[S3Adapter] Retrying to upload part #%v\n", partNumber)
			tryNum++
		} else {
			log.Debugf("[S3Adapter] Uploaded part #%v\n", partNumber)
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}
	return nil, nil
}

func initBucket(ctx context.Context, bucketName string, region string, client *s3.S3) error {
	_, err := client.HeadBucket(&s3.HeadBucketInput{
		Bucket: &bucketName,
	})

	if err != nil {
		_, err := client.CreateBucket(&s3.CreateBucketInput{
			Bucket: &bucketName,
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: &region,
			},
		})

		if err != nil {
			log.Fatal("[S3Adapter] Failed init bucket")
		} else {
			log.Printf("[S3Adapter] Bucket %v exists and you already own it.", bucketName)
		}
	} else {
		log.Printf("[S3Adapter] Successfully created %s\n", bucketName)
	}

	return nil
}
