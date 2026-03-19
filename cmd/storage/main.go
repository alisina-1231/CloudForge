package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"

	. "github.com/alisina-1231/aws-vm/pkg/helpers"
	"github.com/alisina-1231/aws-vm/pkg/mgmt"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	region := MustGetenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg := BuildAWSConfig(context.Background(), region)
	factory := mgmt.NewStorageFactory(cfg, region)
	fmt.Println("Starting to build AWS resources...")
	stack := factory.CreateStorageStack(context.Background(), region)

	uploadBlobs(stack)
	printPresignedURLs(stack)

	fmt.Println("Press enter to delete the infrastructure.")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	factory.DestroyStorageStack(context.Background(), stack)
}

func uploadBlobs(stack *mgmt.StorageStack) {
	s3Client := stack.ServiceClient()
	bucketName := stack.GetBucketName()
	containerName := "jd-imgs" // Using as prefix/folder concept in S3

	fmt.Printf("Creating a new prefix \"jd-imgs\" in the S3 Bucket...\n")

	fmt.Printf("Reading all files /home/ali/aws-vm/blobs...\n")
	files, err := ioutil.ReadDir("/home/ali/aws-vm/blobs")
	HandleErr(err)

	for _, file := range files {
		key := containerName + "/" + file.Name()
		fmt.Printf("Uploading file %q to s3://%s/%s...\n", file.Name(), bucketName, key)

		osFile := HandleErrWithResult(os.Open(path.Join("/home/ali/aws-vm/blobs", file.Name())))
		defer osFile.Close()

		_, err := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   osFile,
		})
		HandleErr(err)
	}
}

func printPresignedURLs(stack *mgmt.StorageStack) {
	s3Client := stack.ServiceClient()
	bucketName := stack.GetBucketName()
	containerName := "jd-imgs"

	fmt.Printf("\nGenerating readonly links to blobs that expire in 2 hours...\n")
	files := HandleErrWithResult(ioutil.ReadDir("/home/ali/aws-vm/blobs"))

	presignClient := s3.NewPresignClient(s3Client)

	for _, file := range files {
		key := containerName + "/" + file.Name()

		req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = 2 * time.Hour
		})
		HandleErr(err)

		fmt.Println(req.URL)
	}
}
