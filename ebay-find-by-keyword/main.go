package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/matthewdargan/swippy-api/ebay-find-by-keyword/finding"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("failed to create AWS SDK session: %w", err)
	}
	ssmClient := ssm.New(sess)
	stage := os.Getenv("STAGE")
	appIDParamName := fmt.Sprintf("/%s/ebay-app-id", stage)
	appID, err := ssmParameterValue(ssmClient, appIDParamName)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("failed to retrieve app ID: %w", err)
	}
	keywords := request.PathParameters["keywords"]
	items, err := finding.FindItemsByKeywords(keywords, appID)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("failed to find eBay items by keywords: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(items),
	}, nil
}

// ssmParameterValue retrieves a parameter value from the AWS SSM Parameter Store.
func ssmParameterValue(ssmClient *ssm.SSM, paramName string) (string, error) {
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve parameter value: %w", err)
	}

	return *output.Parameter.Value, nil
}

func main() {
	lambda.Start(handler)
}
