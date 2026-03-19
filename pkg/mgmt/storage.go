package mgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/yelinaung/go-haikunator"

	. "github.com/alisina-1231/aws-vm/pkg/helpers"
)

var (
	haikuStorage = haikunator.New(time.Now().UnixNano())
)

type StorageStack struct {
	Location   string
	name       string
	BucketName string
	Region     string
	cfg        aws.Config
	s3Client   *s3.Client
}

type StorageFactory struct {
	cfg    aws.Config
	region string
}

// NewStorageFactory instantiates an AWS S3 factory
func NewStorageFactory(cfg aws.Config, region string) *StorageFactory {
	return &StorageFactory{
		cfg:    cfg,
		region: region,
	}
}

func (sf *StorageFactory) CreateStorageStack(ctx context.Context, location string) *StorageStack {
	stack := &StorageStack{
		name:     haikuStorage.Haikunate(),
		Location: location,
		Region:   sf.region,
		cfg:      sf.cfg,
		s3Client: s3.NewFromConfig(sf.cfg),
	}

	// Clean name for S3 (no uppercase, no underscores, 3-63 chars)
	stack.BucketName = strings.ReplaceAll(stack.name, "_", "-") + "-bucket"

	stack.createBucket(ctx)
	return stack
}

func (sf *StorageFactory) DestroyStorageStack(ctx context.Context, stack *StorageStack) {
	fmt.Printf("Deleting S3 bucket %s...\n", stack.BucketName)

	// Delete all objects first
	listRes := HandleErrWithResult(stack.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(stack.BucketName),
	}))

	for _, obj := range listRes.Contents {
		_, err := stack.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(stack.BucketName),
			Key:    obj.Key,
		})
		HandleErr(err)
	}

	// Delete bucket
	_, err := stack.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(stack.BucketName),
	})
	HandleErr(err)
}

func (ss *StorageStack) createBucket(ctx context.Context) {
	fmt.Printf("Building an AWS S3 Bucket named %q...\n", ss.BucketName)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(ss.BucketName),
	}

	// For regions other than us-east-1, we need LocationConstraint
	if ss.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(ss.Region),
		}
	}

	HandleErrWithResult(ss.s3Client.CreateBucket(ctx, input))

	// Wait for bucket to exist
	waiter := s3.NewBucketExistsWaiter(ss.s3Client)
	err := waiter.Wait(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(ss.BucketName),
	}, 2*time.Minute)
	HandleErr(err)
}

func (ss *StorageStack) ServiceClient() *s3.Client {
	return ss.s3Client
}

func (ss *StorageStack) GetBucketName() string {
	return ss.BucketName
}
