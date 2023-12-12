package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/matthewdargan/ebay"
	"github.com/matthewdargan/swippy-api/awsutil"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ssmClient := awsutil.SSMClient()
	appID, err := awsutil.SSMParameterValue(ssmClient, "ebay-app-id")
	if err != nil {
		return errorResponse(http.StatusInternalServerError, fmt.Errorf("failed to retrieve app ID: %w", err))
	}
	fc := ebay.NewFindingClient(&http.Client{Timeout: time.Second * awsutil.FindingHTTPTimeout}, appID)
	resp, err := fc.FindItemsAdvanced(context.Background(), request.QueryStringParameters)
	if err != nil {
		var ebayErr *ebay.APIError
		if errors.As(err, &ebayErr) {
			return errorResponse(ebayErr.StatusCode, ebayErr)
		}
		return errorResponse(http.StatusInternalServerError, err)
	}
	body, err := json.Marshal(resp)
	if err != nil {
		return errorResponse(http.StatusInternalServerError, fmt.Errorf("failed to marshal eBay response: %w", err))
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func errorResponse(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	log.Printf("error: %v", err)
	resp := map[string]string{"error": err.Error()}
	body, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: statusCode}, fmt.Errorf("failed to marshal error: %w", err)
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
