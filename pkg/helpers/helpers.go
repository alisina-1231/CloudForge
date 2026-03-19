package helpers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
)

const (
	DefaultPollingFreq = 10 * time.Second
)

// MustGetenv fetches env variable or exits if missing
func MustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("please add your %s to the .env file", key)
	}
	return val
}

// BuildAWSConfig loads AWS configuration with region
func BuildAWSConfig(ctx context.Context, region string) aws.Config {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	HandleErr(err)
	return cfg
}

// Generic helper for handling errors with return values
func HandleErrWithResult[T any](result T, err error) T {
	HandleErr(err)
	return result
}

// HandleErr processes errors and extracts AWS API errors if present
func HandleErr(err error) {
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("AWS API Error: %s - %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		panic(err)
	}
}

// WaitForResource polls until a condition is met or timeout occurs
func WaitForResource(ctx context.Context, checkFunc func() (bool, error), timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ready, err := checkFunc()
		if err != nil {
			HandleErr(err)
		}

		if ready {
			return
		}

		time.Sleep(DefaultPollingFreq)
	}

	HandleErr(fmt.Errorf("timeout waiting for resource"))
}
