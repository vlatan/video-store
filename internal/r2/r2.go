package r2

import (
	"context"
	"factual-docs/internal/config"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// New creates a new R2 client
func New(ctx context.Context, cfg *config.Config) *s3.Client {

	// Create SDK config for an R2 service
	// An ordinary AWS SDK config would look like:
	// sdkConfig, err := awsConfig.LoadDefaultConfig(ctx)
	sdkConfig, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.R2AccessKeyId, cfg.R2SecretAccessKey, ""),
		),
		awsConfig.WithRegion("auto"),
	)

	if err != nil {
		log.Fatalf("failed to load AWS/R2 SDK configuration, %v", err)
	}

	// Create the R2 client
	// An ordinary AWS client would look like:
	// client := s3.NewFromConfig(sdkConfig)
	return s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		baseEndpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountId)
		o.BaseEndpoint = aws.String(baseEndpoint)
	})
}
