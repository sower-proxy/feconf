package nacos

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/sower-proxy/feconf/reader"
)

func TestParseNacosURI(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		wantErr   bool
		wantGroup string
		wantData  string
	}{
		{
			name:      "valid URI",
			uri:       "nacos://127.0.0.1:8848/DEFAULT_GROUP/app.yaml?namespace=dev&username=user&password=pass&timeout=5s",
			wantGroup: "DEFAULT_GROUP",
			wantData:  "app.yaml",
		},
		{
			name:      "valid URI with escaped slash in dataId",
			uri:       "nacos://127.0.0.1/DEFAULT_GROUP/app%2Fconfig.yaml",
			wantGroup: "DEFAULT_GROUP",
			wantData:  "app/config.yaml",
		},
		{
			name:    "missing group",
			uri:     "nacos://127.0.0.1/app.yaml",
			wantErr: true,
		},
		{
			name:    "missing host",
			uri:     "nacos:///DEFAULT_GROUP/app.yaml",
			wantErr: true,
		},
		{
			name:    "invalid timeout",
			uri:     "nacos://127.0.0.1/DEFAULT_GROUP/app.yaml?timeout=bad",
			wantErr: true,
		},
		{
			name:    "invalid listen timeout",
			uri:     "nacos://127.0.0.1/DEFAULT_GROUP/app.yaml?listen_timeout=0s",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := reader.ParseURI(tt.uri)
			if err != nil {
				t.Fatalf("ParseURI() error = %v", err)
			}

			config, err := parseNacosURI(u)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseNacosURI() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if config.Group != tt.wantGroup {
				t.Errorf("Group = %s, want %s", config.Group, tt.wantGroup)
			}
			if config.DataID != tt.wantData {
				t.Errorf("DataID = %s, want %s", config.DataID, tt.wantData)
			}
			if tt.name == "valid URI" && config.Timeout != 5*time.Second {
				t.Errorf("Timeout = %v, want 5s", config.Timeout)
			}
		})
	}
}

func TestNacosReaderRead(t *testing.T) {
	const testData = `{"name":"app"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != DefaultContextPath+configPath {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("dataId"); got != "app.yaml" {
			t.Errorf("dataId = %s", got)
		}
		if got := r.URL.Query().Get("group"); got != "DEFAULT_GROUP" {
			t.Errorf("group = %s", got)
		}
		if got := r.URL.Query().Get("tenant"); got != "dev" {
			t.Errorf("tenant = %s", got)
		}
		fmt.Fprint(w, testData)
	}))
	defer server.Close()

	nacosReader := newTestReader(t, server.URL+"/DEFAULT_GROUP/app.yaml?namespace=dev")
	data, err := nacosReader.Read(context.Background())
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(data) != testData {
		t.Errorf("Read() data = %s, want %s", string(data), testData)
	}
}

func TestNacosReaderReadWithAuth(t *testing.T) {
	const testData = `{"name":"app"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case DefaultContextPath + authLoginPath:
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if r.Form.Get("username") != "user" || r.Form.Get("password") != "pass" {
				t.Fatalf("unexpected credentials")
			}
			fmt.Fprint(w, `{"accessToken":"token"}`)
		case DefaultContextPath + configPath:
			if got := r.URL.Query().Get("accessToken"); got != "token" {
				t.Errorf("accessToken = %s", got)
			}
			fmt.Fprint(w, testData)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	nacosReader := newTestReader(t, server.URL+"/DEFAULT_GROUP/app.yaml?username=user&password=pass")
	data, err := nacosReader.Read(context.Background())
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(data) != testData {
		t.Errorf("Read() data = %s, want %s", string(data), testData)
	}
}

func TestNacosReaderReadEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	nacosReader := newTestReader(t, server.URL+"/DEFAULT_GROUP/app.yaml")
	_, err := nacosReader.Read(context.Background())
	if err == nil {
		t.Fatal("Read() expected error for empty config")
	}
}

func TestNacosReaderSubscribe(t *testing.T) {
	const testData = `{"name":"app"}`
	listenCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case DefaultContextPath + configListenPath:
			listenCount++
			if got := r.Header.Get("Long-Pulling-Timeout"); got != "10" {
				t.Errorf("Long-Pulling-Timeout = %s", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			payload := r.Form.Get(listeningConfigsParamName)
			if !strings.Contains(payload, "app.yaml"+splitConfigInner+"DEFAULT_GROUP") {
				t.Errorf("listener payload = %q", payload)
			}
			fmt.Fprint(w, "app.yaml")
		case DefaultContextPath + configPath:
			fmt.Fprint(w, testData)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nacosReader := newTestReader(t, server.URL+"/DEFAULT_GROUP/app.yaml?listen_timeout=10ms")

	events, err := nacosReader.Subscribe(ctx)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case event := <-events:
		if !event.IsValid() {
			t.Fatalf("event should be valid: %+v", event)
		}
		if string(event.Data) != testData {
			t.Errorf("event data = %s", string(event.Data))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
	if listenCount == 0 {
		t.Fatal("listener was not called")
	}
}

func TestNacosReaderClose(t *testing.T) {
	nacosReader := newTestReader(t, "http://127.0.0.1:8848/DEFAULT_GROUP/app.yaml")

	if err := nacosReader.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := nacosReader.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if _, err := nacosReader.Read(context.Background()); err == nil {
		t.Fatal("Read() after Close() expected error")
	}
}

func TestNacosReaderCloseStopsSubscription(t *testing.T) {
	block := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
	}))
	defer server.Close()
	defer close(block)

	nacosReader := newTestReader(t, server.URL+"/DEFAULT_GROUP/app.yaml?listen_timeout=5s")
	events, err := nacosReader.Subscribe(context.Background())
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if err := nacosReader.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	timeout := time.After(time.Second)
	for {
		select {
		case _, ok := <-events:
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("timeout waiting for subscription close")
		}
	}
}

func TestNacosReaderInterface(t *testing.T) {
	var _ reader.ConfReader = &NacosReader{}
}

func newTestReader(t *testing.T, rawURL string) *NacosReader {
	t.Helper()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	query := parsedURL.RawQuery
	uri := "nacos://" + parsedURL.Host + parsedURL.Path
	if query != "" {
		uri += "?" + query
	}

	nacosReader, err := NewNacosReader(uri + queryServerScheme(parsedURL.Scheme, query))
	if err != nil {
		t.Fatalf("NewNacosReader() error = %v", err)
	}
	return nacosReader
}

func queryServerScheme(scheme, query string) string {
	separator := "?"
	if query != "" {
		separator = "&"
	}
	return separator + "server_scheme=" + scheme
}
