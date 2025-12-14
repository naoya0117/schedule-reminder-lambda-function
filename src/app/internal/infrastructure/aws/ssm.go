package aws

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMClient is a client for AWS Systems Manager Parameter Store
type SSMClient struct {
	client *ssm.Client
	prefix string
}

// NewSSMClient creates a new SSM client
// It automatically configures for LocalStack when AWS_ENDPOINT_URL is set
func NewSSMClient(ctx context.Context) (*SSMClient, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create an SSM client
	client := ssm.NewFromConfig(cfg, func(o *ssm.Options) {
		// Override endpoint for LocalStack if AWS_ENDPOINT_URL is set
		if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	// Get parameter prefix from environment or use default
	prefix := os.Getenv("SSM_PARAM_PREFIX")
	if prefix == "" {
		prefix = "/lambda-functions/schedule-reminder"
	}

	return &SSMClient{
		client: client,
		prefix: prefix,
	}, nil
}

// GetParameter retrieves a parameter from Parameter Store
// paramName should be the variable name (e.g., "NOTION_API_KEY")
// It will be converted to the full path following the PARAM_ convention:
// "NOTION_API_KEY" -> "/lambda-functions/schedule-reminder/param-notion-api-key"
func (c *SSMClient) GetParameter(ctx context.Context, paramName string) (string, error) {
	// Convert parameter name to Parameter Store path
	// Following the convention: PARAM_NOTION_API_KEY -> param-notion-api-key
	paramPath := c.buildParameterPath(paramName)

	input := &ssm.GetParameterInput{
		Name:           aws.String(paramPath),
		WithDecryption: aws.Bool(true),
	}

	result, err := c.client.GetParameter(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get parameter %s: %w", paramPath, err)
	}

	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", fmt.Errorf("parameter %s has no value", paramPath)
	}

	return *result.Parameter.Value, nil
}

// GetParameterWithFallback retrieves a parameter from Parameter Store,
// falling back to an environment variable if not found
func (c *SSMClient) GetParameterWithFallback(ctx context.Context, paramName string) (string, error) {
	// Try to get from Parameter Store
	value, err := c.GetParameter(ctx, paramName)
	if err == nil {
		return value, nil
	}

	// Fallback to environment variable
	envValue := os.Getenv(paramName)
	if envValue != "" {
		fmt.Printf("Warning: Using environment variable %s as fallback (Parameter Store: %v)\n", paramName, err)
		return envValue, nil
	}

	return "", fmt.Errorf("parameter %s not found in Parameter Store or environment: %w", paramName, err)
}

// buildParameterPath converts a parameter name to the full Parameter Store path
// Following the PARAM_ convention from docker-compose:
// "NOTION_API_KEY" -> "/lambda-functions/schedule-reminder/param-notion-api-key"
func (c *SSMClient) buildParameterPath(paramName string) string {
	// Convert to lowercase and replace underscores with hyphens
	paramKey := strings.ToLower(strings.ReplaceAll(paramName, "_", "-"))
	
	// Add "param-" prefix
	paramKey = "param-" + paramKey
	
	// Combine with prefix
	return fmt.Sprintf("%s/%s", c.prefix, paramKey)
}
