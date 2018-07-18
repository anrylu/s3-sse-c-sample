package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	SSECustomerAlgorithm = "AES256"
)

type QS3Config struct {
	UseAccessKey            bool
	AccessKeyID             string
	SecretAccessKey         string
	Region                  string
	BucketName              string
	SSECustomerKey          string
	SSECustomerKeyBase64    string
	SSECustomerKeyMD5Base64 string
}

type QS3 struct {
	Cfg   *QS3Config `inject:""`
	S3Svc *s3.S3
}

func (qs3 *QS3) Setup(ctx context.Context) {
	// init s3
	ec2m := ec2metadata.New(session.New(), &aws.Config{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})
	cr := credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{
		Client: ec2m,
	})
	if qs3.Cfg.UseAccessKey {
		token := ""
		log.Println(qs3.Cfg.AccessKeyID)
		log.Println(qs3.Cfg.SecretAccessKey)
		cr = credentials.NewStaticCredentials(qs3.Cfg.AccessKeyID, qs3.Cfg.SecretAccessKey, token)
	}
	s3cfg := aws.NewConfig()
	s3cfg = s3cfg.WithRegion(qs3.Cfg.Region).WithCredentials(cr)
	s3Session := session.Must(session.NewSession(s3cfg))
	qs3.S3Svc = s3.New(s3Session)

	// get SSECustomerKey
	data, err := base64.StdEncoding.DecodeString(qs3.Cfg.SSECustomerKeyBase64)
	if err != nil {
		log.Fatal(ctx, "invalid SSECustomerKey", zap.Error(err))
	}
	qs3.Cfg.SSECustomerKey = string(data)
}

func (qs3 *QS3) Upload(c *gin.Context) error {
	// get data
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	defer src.Close()
	if err != nil {
		return err
	}

	// init multipart upload
	log.Println("init multipart upload")
	createMultipartUploadInput := s3.CreateMultipartUploadInput{
		Bucket:               aws.String(qs3.Cfg.BucketName),
		Key:                  aws.String(file.Filename),
		SSECustomerAlgorithm: aws.String(SSECustomerAlgorithm),
		SSECustomerKey:       aws.String(qs3.Cfg.SSECustomerKey),
		SSECustomerKeyMD5:    aws.String(qs3.Cfg.SSECustomerKeyMD5Base64),
	}
	initOut, err := qs3.S3Svc.CreateMultipartUploadWithContext(c, &createMultipartUploadInput)
	if err != nil {
		log.Println("init multipart upload failed", zap.Error(err))
		return err
	}

	// upload file content
	var partNumber int64 = 1
	completeParts := []*s3.CompletedPart{}
	b := make([]byte, 1024*1024*5)
	for {
		log.Printf("upload part %d\n", partNumber)
		n, err := src.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			qs3.S3Svc.AbortMultipartUploadWithContext(c, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(qs3.Cfg.BucketName),
				Key:      aws.String(file.Filename),
				UploadId: initOut.UploadId,
			})
			return err
		}
		uploadPartInput := s3.UploadPartInput{
			Bucket:               aws.String(qs3.Cfg.BucketName),
			Key:                  aws.String(file.Filename),
			UploadId:             initOut.UploadId,
			PartNumber:           aws.Int64(partNumber),
			Body:                 bytes.NewReader(b[:n]),
			SSECustomerAlgorithm: aws.String(SSECustomerAlgorithm),
			SSECustomerKey:       aws.String(qs3.Cfg.SSECustomerKey),
			SSECustomerKeyMD5:    aws.String(qs3.Cfg.SSECustomerKeyMD5Base64),
		}
		uploadOut, err := qs3.S3Svc.UploadPartWithContext(c, &uploadPartInput)
		if err != nil {
			log.Println("multipart upload failed", zap.Error(err))
			qs3.S3Svc.AbortMultipartUploadWithContext(c, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(qs3.Cfg.BucketName),
				Key:      aws.String(file.Filename),
				UploadId: initOut.UploadId,
			})
			return err
		}
		completeParts = append(completeParts, &s3.CompletedPart{
			ETag:       uploadOut.ETag,
			PartNumber: aws.Int64(partNumber),
		})
		partNumber = partNumber + 1
	}

	// complete multipart upload
	log.Println("complete multipart upload")
	_, err = qs3.S3Svc.CompleteMultipartUploadWithContext(c, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(qs3.Cfg.BucketName),
		Key:      aws.String(file.Filename),
		UploadId: initOut.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completeParts,
		},
	})

	c.String(http.StatusOK, "upload successfully!")
	return nil
}

func (qs3 *QS3) Download(c *gin.Context, filename string) error {
	// get object
	obj, err := qs3.S3Svc.GetObjectWithContext(c, &s3.GetObjectInput{
		Bucket:               aws.String(qs3.Cfg.BucketName),
		Key:                  aws.String(filename),
		SSECustomerAlgorithm: aws.String(SSECustomerAlgorithm),
		SSECustomerKey:       aws.String(qs3.Cfg.SSECustomerKey),
		SSECustomerKeyMD5:    aws.String(qs3.Cfg.SSECustomerKeyMD5Base64),
	})
	if err != nil {
		log.Printf("get object %s failed", filename, zap.Error(err))
		return err
	}

	// set header
	c.Writer.Header().Set("Content-Length", strconv.FormatInt(*obj.ContentLength, 10))

	// read body and output
	var b = make([]byte, 8192)
	for {
		n, err := obj.Body.Read(b)
		if n > 0 {
			c.Writer.Write(b[:n])
		}
		if err != nil {
			break
		}
		runtime.Gosched()
	}

	return nil
}
