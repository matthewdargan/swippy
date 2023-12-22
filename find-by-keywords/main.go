// find-by-keywords is an AWS Lambda that requests the eBay Finding API findItemsByKeywords endpoint.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/matthewdargan/swippy-api/ebay"
)

const eBayParamName = "ebay-app-id"

var client = &http.Client{
	Timeout: time.Second * 10,
}

func handleRequest(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("failed to create AWS SDK session: %w", err)
	}
	ssmClient := ssm.New(sess)
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(eBayParamName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("failed to retrieve parameter value: %w", err)
	}
	findingClient := ebay.NewFindingClient(client, *output.Parameter.Value)
	resp, err := findingClient.FindItemsByKeywords(req.QueryStringParameters)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("failed to marshal eBay API response: %w", err)
	}
	if len(resp.ItemsResponse) > 0 && len(resp.ItemsResponse[0].ErrorMessage) > 0 {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(data),
		}, nil
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(data),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
