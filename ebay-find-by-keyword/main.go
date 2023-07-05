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
	"github.com/matthewdargan/swippy-api/ebay-find-by-keyword/ebay"
)

const findingHTTPTimeout = 5

var (
	ErrKeywordsMissing = errors.New("keywords parameter is required")
	stage              string
	findingClient      *http.Client
)

func init() {
	stage = os.Getenv("STAGE")
	findingClient = &http.Client{
		Timeout: time.Second * findingHTTPTimeout,
	}
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	keywords, found := request.QueryStringParameters["keywords"]
	if !found {
		return generateErrorResponse(http.StatusBadRequest, ErrKeywordsMissing)
	}

	findingParams := &ebay.FindingParams{
		Keywords: keywords,
	}
	aspectName, anFound := request.QueryStringParameters["aspectFilter.aspectName"]
	aspectValueName, anvFound := request.QueryStringParameters["aspectFilter.aspectValueName"]
	// TODO: Stricten this so that it 404s if only 1 of the 2 filter parts are passed in
	if anFound && anvFound {
		findingParams.AspectFilter = &ebay.AspectFilter{
			AspectName:      aspectName,
			AspectValueName: aspectValueName,
		}
	}

	ssmClient := ssmClient()
	appIDParamName := fmt.Sprintf("/%s/ebay-app-id", stage)
	appID, err := ssmParameterValue(ssmClient, appIDParamName)
	if err != nil {
		return generateErrorResponse(http.StatusInternalServerError, fmt.Errorf("failed to retrieve app ID: %w", err))
	}

	findingSvr := ebay.NewFindingServer(findingClient)
	items, err := findingSvr.FindItemsByKeywords(findingParams, appID)
	if err != nil {
		return generateErrorResponse(
			http.StatusInternalServerError, fmt.Errorf("failed to find eBay items by keywords: %w", err))
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
