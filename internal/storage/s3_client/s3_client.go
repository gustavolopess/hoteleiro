package s3_client

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gustavolopess/hoteleiro/internal/config"
	"golang.org/x/oauth2"
)

type S3Client struct {
	*session.Session
	*s3manager.Downloader
}

var client *S3Client

func initS3Client() *S3Client {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(config.AwsRegion)},
	)

	downloader := s3manager.NewDownloader(sess)

	return &S3Client{
		Session:    sess,
		Downloader: downloader,
	}
}

func GetS3Client() *S3Client {
	if client == nil {
		return initS3Client()
	}
	return client
}

func (c *S3Client) GetGoogleSheetsCreds() []byte {
	item := config.GoogleSheetsCredentialsInS3

	b := make([]byte, 1024)
	buf := aws.NewWriteAtBuffer(b)
	numBytes, err := c.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(config.S3Bucket),
			Key:    aws.String(item),
		})

	if err != nil {
		log.Fatalf("Unable to download item %q, %v", item, err)
	}

	log.Println("Downloaded", item, numBytes, "bytes")

	return b[:numBytes]
}

func (c *S3Client) GetGoogleSheetsAuthToken() *oauth2.Token {
	item := config.GoogleSheetsTokenInS3

	b := make([]byte, 1024)
	buf := aws.NewWriteAtBuffer(b)
	numBytes, err := c.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(config.S3Bucket),
			Key:    aws.String(item),
		})

	if err != nil {
		log.Fatalf("Unable to download item %q, %v", item, err)
	}

	log.Println("Downloaded", item, numBytes, "bytes")

	tok := &oauth2.Token{}
	err = json.Unmarshal(b[:numBytes], tok)
	if err != nil {
		log.Fatalf("Failed to read json of %q, %v", item, err)
	}

	return tok
}
