package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/matthewdargan/swippy-api/ebay"
)

const findingHTTPTimeout = 5

var (
	stage  string
	client *http.Client
)

func init() {
	stage = os.Getenv("STAGE")
	client = &http.Client{
		Timeout: time.Second * findingHTTPTimeout,
	}
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ssmClient := ssmClient()
	appIDParamName := fmt.Sprintf("/%s/ebay-app-id", stage)
	appID, err := ssmParameterValue(ssmClient, appIDParamName)
	if err != nil {
		return generateErrorResponse(http.StatusInternalServerError, fmt.Errorf("failed to retrieve app ID: %w", err))
	}

	fc := ebay.NewFindingClient(client, appID)
	items, err := fc.FindItemsByCategories(request.QueryStringParameters)
	if err != nil {
		var ebayErr *ebay.APIError
		if errors.As(err, &ebayErr) {
			return generateErrorResponse(ebayErr.StatusCode, ebayErr)
		}

		return generateErrorResponse(http.StatusInternalServerError, err)
	}
	if err != nil {
		return generateErrorResponse(
			http.StatusInternalServerError, fmt.Errorf("failed to find eBay items by categories: %w", err))
	}

	body, err := json.Marshal(items)
	if err != nil {
		return generateErrorResponse(
			http.StatusInternalServerError, fmt.Errorf("failed to marshal eBay items response: %w", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func ssmClient() *ssm.SSM {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to create AWS SDK session: %v", err)
	}

	return ssm.New(sess)
}

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

func generateErrorResponse(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	log.Printf("error: %v", err)
	resp := map[string]string{
		"error": err.Error(),
	}
	body, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: statusCode}, fmt.Errorf("failed to marshal error response: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
