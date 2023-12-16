// find-by-product is an AWS Lambda that requests the eBay Finding API findItemsByProduct endpoint.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/matthewdargan/swippy-api/internal/ebay"
)

var (
	client = http.Client{Timeout: 10 * time.Second}
	config = ebay.EBayConfig{
		AppID:              "ebay-app-id",
		BaseURL:            "https://svcs.ebay.com/services/search/FindingService/v1",
		OperationName:      "findItemsByProduct",
		ServiceVersion:     "1.0.0",
		ResponseDataFormat: "JSON",
		ContentType:        "application/json",
	}
)

type apiError struct {
	StatusCode int
	Message    error
}

func (e *apiError) Error() string {
	return fmt.Sprintf("API request failed with status code %d: %v", e.StatusCode, e.Message)
}

func errorResponse(code int, err error) (events.APIGatewayProxyResponse, *apiError) {
	if err == nil {
		err = errors.New("unknown error")
	}
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Headers:    map[string]string{"Content-Type": config.ContentType},
		Body:       err.Error(),
	}, &apiError{StatusCode: code, Message: err}
}

// TODO: events.APIGatewayV2HTTPRequest and events.APIGatewayV2HTTPResponse
func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	apiReq, err := config.APIRequest(ctx, req)
	if err != nil {
		return errorResponse(http.StatusInternalServerError, err)
	}
	resp, err := client.Do(apiReq)
	if err != nil {
		return errorResponse(resp.StatusCode, fmt.Errorf("failed to make eBay API request: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errorResponse(resp.StatusCode, fmt.Errorf("eBay API request failed"))
	}
	var res ebay.FindItemsByProductResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return errorResponse(http.StatusInternalServerError, fmt.Errorf("failed to unmarshal eBay API response: %w", err))
	}
	body, err := json.Marshal(res)
	if err != nil {
		return errorResponse(http.StatusInternalServerError, fmt.Errorf("failed to marshal eBay API response: %w", err))
	}
	if len(res.ItemsResponse) > 0 && len(res.ItemsResponse[0].ErrorMessage) > 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    map[string]string{"Content-Type": config.ContentType},
			Body:       string(body),
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: resp.StatusCode,
		Headers:    map[string]string{"Content-Type": config.ContentType},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
