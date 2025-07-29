package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/veteran-software/yourddo-api/shared/types"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

var allowedMethods = []string{http.MethodGet, http.MethodOptions}

// handleRequest handles incoming API Gateway requests to fetch server statuses and return responses with proper CORS headers.
func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	method := strings.ToUpper(req.HTTPMethod)
	if method == "" {
		method = http.MethodGet
	}

	if strings.EqualFold(method, http.MethodOptions) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      corsOrigin(req.Path),
				"Access-Control-Allow-Methods":     strings.Join(allowedMethods, ","),
				"Access-Control-Allow-Headers":     "Content-Type,Authorization",
				"Access-Control-Allow-Credentials": "true",
			},
		}, nil
	}

	if !strings.EqualFold(method, http.MethodGet) {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusMethodNotAllowed}, nil
	}

	servers, errors := fetchServerStatus(ctx)

	errorStrings := make([]string, 0, len(errors))
	for _, err := range errors {
		errorStrings = append(errorStrings, err.Error())
	}

	response := types.Response{
		Servers: servers,
		Errors:  errorStrings,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": corsOrigin(req.Path),
		},
		Body: string(jsonResponse),
	}, nil
}

// FetchAndParseDatacenter fetches XML data from the given URL and parses it into a structured ArrayOfDatacenterStruct.
// Returns the parsed data or an error if the request fails or the data cannot be parsed.
func FetchAndParseDatacenter(url string) (*types.ArrayOfDatacenterStruct, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return ParseDatacenterXML(resp.Body)
}

// FetchAndParseStatus retrieves an XML status document from the given URL, parses it, and returns a Status struct.
// Returns an error if the request fails, the response cannot be read, or parsing fails.
// Retrieves and decodes the XML into the types.Status struct while ensuring proper charset handling.
func FetchAndParseStatus(url string) (*types.Status, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch status: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	body = bytes.TrimSpace(body)

	var status types.Status
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charset.NewReaderLabel // Handle potential charset issues

	if err := decoder.Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return &status, nil
}

// fetchServerStatus retrieves server information and status from a datacenter URL and returns a list of servers with errors.
func fetchServerStatus(ctx context.Context) ([]*types.ServerInfo, []error) {
	url := os.Getenv("DATACENTER_URL")
	if url == "" {
		return nil, []error{fmt.Errorf("DATACENTER_URL environment variable is not set")}
	}

	result, err := FetchAndParseDatacenter(url)
	if err != nil {
		return nil, []error{err}
	}

	worldInfo := make(map[string]*types.World)
	var urls []string
	for _, world := range result.DatacenterStructs[0].Datacenter.Datacenter.Worlds {
		worldInfo[world.StatusServerUrl] = &world
		urls = append(urls, world.StatusServerUrl)
	}

	pool := NewWorkerPool(len(urls), 0)
	results := pool.ProcessURLs(ctx, urls)

	serverInfos := make([]*types.ServerInfo, 0, len(urls))
	var errors []error

	for result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("URL %s: %w", result.URL, result.Error))
			continue
		}

		if result.Status == nil {
			continue
		}

		world := worldInfo[result.URL]

		isActive := true
		if result.Status.AllowBillingRole == "" {
			isActive = false
		} else {
			roles := strings.Split(result.Status.AllowBillingRole, ",")
			isActive = len(roles) >= 5
		}

		serverInfos = append(serverInfos, &types.ServerInfo{
			Name:       world.Name,
			CommonName: result.Status.Name,
			Status:     isActive,
			Order:      world.Order,
		})
	}

	sort.Slice(serverInfos, func(i, j int) bool {
		return serverInfos[i].Order < serverInfos[j].Order
	})

	return serverInfos, errors
}

// corsOrigin determines the CORS origin URL based on the provided path. Returns a specific URL for "/server_status".
func corsOrigin(path string) string {
	if path == "/server_status" {
		return "https://ddocompendium.com"
	}

	return "https://yourddo.com"
}

// main is the entry point of the application, initializing the Lambda function and starting the request handler.
func main() {
	lambda.Start(handleRequest)
}
