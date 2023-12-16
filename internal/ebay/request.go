package ebay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/matthewdargan/swippy-api/internal/awsutil"
)

type EBayConfig struct {
	AppID              string
	BaseURL            string
	OperationName      string
	ServiceVersion     string
	ResponseDataFormat string
	ContentType        string
}

func (c EBayConfig) APIRequest(ctx context.Context, req events.APIGatewayProxyRequest) (*http.Request, error) {
	ssmClient := awsutil.SSMClient()
	appID, err := awsutil.SSMParameterValue(ssmClient, c.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve app ID: %w", err)
	}
	url := fmt.Sprintf("%s?%s", c.BaseURL, c.queryString(req, appID))
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c EBayConfig) queryString(req events.APIGatewayProxyRequest, appID string) string {
	qry := url.Values{
		"Operation-Name":       {c.OperationName},
		"Service-Version":      {c.ServiceVersion},
		"Security-AppName":     {appID},
		"Response-Data-Format": {c.ResponseDataFormat},
		"REST-Payload":         {""},
	}
	for k, v := range req.QueryStringParameters {
		if v != "" {
			qry.Set(k, v)
		}
	}
	return qry.Encode()
}
