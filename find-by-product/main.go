// find-by-product is an AWS Lambda that requests the eBay Finding API findItemsByProduct endpoint.
package main

import (
	"encoding/json"
	"log"
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
		log.Println("failed to create AWS SDK session:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	ssmClient := ssm.New(sess)
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(eBayParamName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Println("failed to retrieve parameter value:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	findingClient := ebay.NewFindingClient(client, *output.Parameter.Value)
	resp, err := findingClient.FindItemsByProduct(req.QueryStringParameters)
	if err != nil {
		log.Println("failed to execute eBay API request:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("failed to marshal eBay API response:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
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
