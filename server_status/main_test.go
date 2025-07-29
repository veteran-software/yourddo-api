package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/veteran-software/yourddo-api/shared/types"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

const contentTypeKey = "Content-Type"
const contentTypeValue = "application/xml"
const DatacenterUrl = "DATACENTER_URL"
const invalidUrl = "http://invalid-url"

// Helper function to create a test server with static XML response
func newXMLTestServer(responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeKey, contentTypeValue)
		_, err := w.Write([]byte(responseBody))
		if err != nil {
			panic(err)
		}
	}))
}

var statusResponse = `
    <Status>
        <name>TestServer</name>
        <allow_billing_role>role1,role2,role3,role4,role5</allow_billing_role>
    </Status>`

var dcResponse = `
    <ArrayOfDatacenterStruct>
        <DatacenterStruct>
            <KeyName>Test</KeyName>
            <Datacenter>
                <cachedAt>2024-01-01T00:00:00Z</cachedAt>
                <datacenter>
                    <Datacenter>
                        <Worlds>
                            <World>
                                <Name>TestWorld</Name>
                                <StatusServerUrl>%s</StatusServerUrl>
                                <Order>1</Order>
                            </World>
                        </Worlds>
                    </Datacenter>
                </datacenter>
            </Datacenter>
        </DatacenterStruct>
    </ArrayOfDatacenterStruct>`

func setEnvVar(key, value string) error {
	return os.Setenv(key, value)
}

func unsetEnvVar(key string) error {
	return os.Unsetenv(key)
}

func setupEnv(t *testing.T, envURL string) func() {
	if envURL != "" {
		if err := setEnvVar(DatacenterUrl, envURL); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := unsetEnvVar(DatacenterUrl); err != nil {
			t.Fatal(err)
		}
	}
	return func() {
		if envURL != "" {
			_ = unsetEnvVar(DatacenterUrl)
		}
	}
}

func TestHandleRequest(t *testing.T) {
	// Setup test servers
	statusServer := newXMLTestServer(statusResponse)
	defer statusServer.Close()

	datacenterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeKey, contentTypeValue)
		_, err := w.Write([]byte(fmt.Sprintf(dcResponse, statusServer.URL)))
		if err != nil {
			return
		}
	}))
	defer datacenterServer.Close()

	tests := []struct {
		name          string
		envURL        string
		wantStatus    int
		wantErrorsLen int
	}{
		{
			name:          "missing environment variable",
			envURL:        "",
			wantStatus:    200,
			wantErrorsLen: 1,
		},
		{
			name:          "valid datacenter URL",
			envURL:        datacenterServer.URL,
			wantStatus:    200,
			wantErrorsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupEnv(t, tt.envURL)
			defer cleanup()

			got, err := handleRequest(context.Background(), events.APIGatewayProxyRequest{})
			if err != nil {
				t.Errorf("handleRequest() error = %v", err)
				return
			}

			if got.StatusCode != tt.wantStatus {
				t.Errorf("status = %v, want %v", got.StatusCode, tt.wantStatus)
			}

			var response types.Response
			if err := json.Unmarshal([]byte(got.Body), &response); err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
				return
			}

			if len(response.Errors) != tt.wantErrorsLen {
				t.Errorf("errors length = %v, want %v", len(response.Errors), tt.wantErrorsLen)
			}
		})

	}
}

func TestFetchServerStatus(t *testing.T) {
	// Setup test servers
	statusServer := newXMLTestServer(statusResponse)
	defer statusServer.Close()

	datacenterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeKey, contentTypeValue)
		_, err := w.Write([]byte(fmt.Sprintf(dcResponse, statusServer.URL)))
		if err != nil {
			return
		}
	}))
	defer datacenterServer.Close()

	tests := []struct {
		name        string
		envURL      string
		wantServers int
		wantErrors  int
	}{
		{
			name:        "missing environment variable",
			envURL:      "",
			wantServers: 0,
			wantErrors:  1,
		},
		{
			name:        "valid datacenter URL",
			envURL:      datacenterServer.URL,
			wantServers: 1,
			wantErrors:  0,
		},
		{
			name:        "invalid datacenter URL",
			envURL:      invalidUrl,
			wantServers: 0,
			wantErrors:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}
func TestCorsOrigin(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "server status path",
			path:     "/server_status",
			expected: "https://ddocompendium.com",
		},
		{
			name:     "other path",
			path:     "/random_path",
			expected: "https://yourddo.com",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "https://yourddo.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := corsOrigin(tt.path)
			if got != tt.expected {
				t.Errorf("corsOrigin(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}
func runTest(t *testing.T, tt struct {
	name        string
	envURL      string
	wantServers int
	wantErrors  int
}) {
	if tt.envURL != "" {
		err := os.Setenv("DATACENTER_URL", tt.envURL)
		if err != nil {
			return
		}
		defer func() {
			err := os.Unsetenv("DATACENTER_URL")
			if err != nil {
				return
			}
		}()
	} else {
		err := os.Unsetenv("DATACENTER_URL")
		if err != nil {
			return
		}
	}

	servers, errors := fetchServerStatus(context.Background())

	checkServers(t, servers, tt)

	checkErrors(t, errors, tt)
}

func checkErrors(t *testing.T, errors []error, tt struct {
	name        string
	envURL      string
	wantServers int
	wantErrors  int
}) {
	if len(errors) != tt.wantErrors {
		t.Errorf("fetchServerStatus() errors = %v, want %v", len(errors), tt.wantErrors)
	}
}

func checkServers(t *testing.T, servers []*types.ServerInfo, tt struct {
	name        string
	envURL      string
	wantServers int
	wantErrors  int
}) {
	if len(servers) != tt.wantServers {
		t.Errorf("fetchServerStatus() servers = %v, want %v", len(servers), tt.wantServers)
	}
}

func TestFetchAndParseDatacenter(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeKey, contentTypeValue)
		_, err := w.Write([]byte(`
			<ArrayOfDatacenterStruct>
				<DatacenterStruct>
					<KeyName>Test</KeyName>
					<Datacenter>
						<cachedAt>2024-01-01T00:00:00Z</cachedAt>
						<datacenter>
							<Datacenter>
								<Name>TestDC</Name>
							</Datacenter>
						</datacenter>
					</Datacenter>
				</DatacenterStruct>
			</ArrayOfDatacenterStruct>`))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     server.URL,
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     invalidUrl,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchAndParseDatacenter(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAndParseDatacenter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("FetchAndParseDatacenter() returned nil for valid URL")
			}
		})
	}
}

func TestFetchAndParseStatus(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeKey, contentTypeValue)
		_, err := w.Write([]byte(statusResponse))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		want    *types.Status
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     server.URL,
			want:    &types.Status{Name: "TestServer", AllowBillingRole: "role1,role2,role3,role4,role5"},
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     invalidUrl,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchAndParseStatus(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAndParseStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("FetchAndParseStatus() returned nil for valid URL")
				} else if !reflect.DeepEqual(got.Name, tt.want.Name) {
					t.Errorf("FetchAndParseStatus() = %v, want %v", got.Name, tt.want.Name)
				}
			}
		})
	}
}
