package server_status

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

func handleRequest(ctx context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	servers, errors := fetchServerStatus(ctx)

	errorStrings := make([]string, 0, len(errors))
	for _, err := range errors {
		errorStrings = append(errorStrings, err.Error())
	}

	response := types.Response{
		Servers: servers,
		Errors:  errorStrings,
	}

	// Convert response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonResponse),
	}, nil
}

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

// noinspection GoUnusedFunction
func main() {
	lambda.Start(handleRequest)
}
