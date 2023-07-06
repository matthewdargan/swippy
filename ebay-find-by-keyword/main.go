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
	ErrKeywordsMissing              = errors.New("keywords parameter is required")
	ErrIncompleteAspectFilter       = errors.New("incomplete aspect filter: aspectName and aspectValueName are required")
	ErrIncompleteItemFilterNameOnly = errors.New("incomplete item filter: missing value")
	ErrIncompleteItemFilterParam    = errors.New("incomplete item filter: missing param value")
	stage                           string
	findingClient                   *http.Client
)

func init() {
	stage = os.Getenv("STAGE")
	findingClient = &http.Client{
		Timeout: time.Second * findingHTTPTimeout,
	}
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	keywords, ok := request.QueryStringParameters["keywords"]
	if !ok {
		return generateErrorResponse(http.StatusBadRequest, ErrKeywordsMissing)
	}

	findingParams := &ebay.FindingParams{
		Keywords: keywords,
	}

	aspectName, anOk := request.QueryStringParameters["aspectFilter.aspectName"]
	aspectValueName, avnOk := request.QueryStringParameters["aspectFilter.aspectValueName"]
	if anOk != avnOk {
		return generateErrorResponse(http.StatusNotFound, ErrIncompleteAspectFilter)
	}
	if anOk && avnOk {
		findingParams.AspectFilter = &ebay.AspectFilter{
			AspectName:      aspectName,
			AspectValueName: aspectValueName,
		}
	}

	for idx := 0; ; idx++ {
		ifName, ok := request.QueryStringParameters[fmt.Sprintf("itemFilter(%d).name", idx)]
		if !ok {
			break
		}

		ifValue, ok := request.QueryStringParameters[fmt.Sprintf("itemFilter(%d).value", idx)]
		if !ok {
			return generateErrorResponse(http.StatusNotFound, ErrIncompleteItemFilterNameOnly)
		}

		itemFilter := ebay.ItemFilter{
			Name:  ifName,
			Value: ifValue,
		}

		ifParamName, pnOk := request.QueryStringParameters[fmt.Sprintf("itemFilter(%d).paramName", idx)]
		ifParamValue, pvOk := request.QueryStringParameters[fmt.Sprintf("itemFilter(%d).paramValue", idx)]
		if pnOk != pvOk {
			return generateErrorResponse(http.StatusNotFound, ErrIncompleteItemFilterParam)
		}
		if pnOk && pvOk {
			itemFilter.ParamName = &ifParamName
			itemFilter.ParamValue = &ifParamValue
		}

		findingParams.ItemFilters = append(findingParams.ItemFilters, itemFilter)
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
