// Copyright 2024 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

// find-by-product is an AWS Lambda that requests the eBay Finding API findItemsByProduct endpoint.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/matthewdargan/ebay"
)

const (
	eBayParamName = "ebay-app-id"
	sqsQueueName  = "swippy-api-queue"
)

var client = &http.Client{
	Timeout: time.Second * 10,
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("failed to create AWS SDK session:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	ssmSvc := ssm.New(sess)
	paramRes, err := ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(eBayParamName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Println("failed to retrieve parameter value:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	findingClient := ebay.NewFindingClient(client, *paramRes.Parameter.Value)
	resp, err := findingClient.FindItemsByProduct(ctx, req.QueryStringParameters)
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
	itemsResponseData, err := json.Marshal(resp.ItemsResponse[0])
	if err != nil {
		log.Println("failed to marshal eBay API response item:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	sqsSvc := sqs.New(sess)
	urlRes, err := sqsSvc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(sqsQueueName),
	})
	if err != nil {
		log.Println("failed to retrieve SQS queue URL:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	_, err = sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(itemsResponseData)),
		QueueUrl:    urlRes.QueueUrl,
	})
	if err != nil {
		log.Println("failed to send message to SQS queue:", err)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
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
