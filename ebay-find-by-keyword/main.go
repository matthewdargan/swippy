package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/matthewdargan/swippy-api/ebay-find-by-keyword/finding"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	keywords := request.PathParameters["keywords"]
	respBody, err := finding.FindItemsByKeywords(keywords)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("failed to find eBay items by keywords: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(respBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
