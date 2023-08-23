package awsutil

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const FindingHTTPTimeout = 5

func SSMClient() *ssm.SSM {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("awsutil: failed to create AWS SDK session: %v", err)
	}
	return ssm.New(sess)
}

func SSMParameterValue(ssmClient *ssm.SSM, paramName string) (string, error) {
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("awsutil: failed to retrieve parameter value: %w", err)
	}
	return *output.Parameter.Value, nil
}
