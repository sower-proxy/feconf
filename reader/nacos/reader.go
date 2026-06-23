package nacos

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sower-proxy/feconf/reader"
)

const (
	// SchemeNacos represents Nacos URI scheme.
	SchemeNacos reader.Scheme = "nacos"

	DefaultPort               = 8848
	DefaultTimeout            = 10 * time.Second
	DefaultListenTimeout      = 30 * time.Second
	DefaultRetryDelay         = time.Second
	DefaultContextPath        = "/nacos"
	splitConfig               = "\x01"
	splitConfigInner          = "\x02"
	configPath                = "/v1/cs/configs"
	configListenPath          = "/v1/cs/configs/listener"
	authLoginPath             = "/v1/auth/users/login"
	listeningConfigsParamName = "Listening-Configs"
)

func init() {
	_ = reader.RegisterReader(SchemeNacos, func(uri string) (reader.ConfReader, error) {
		return NewNacosReader(uri)
	})
}

// NacosConfig holds Nacos reader configuration.
type NacosConfig struct {
	BaseURL       string
	Group         string
	DataID        string
	Namespace     string
	Username      string
	Password      string
	Timeout       time.Duration
	ListenTimeout time.Duration
	RetryDelay    time.Duration
}

// NacosReader implements ConfReader for Nacos configuration.
type NacosReader struct {
	uri         string
	config      *NacosConfig
	client      *http.Client
	closeCtx    context.Context
	closeCancel context.CancelFunc
	accessToken string
	tokenMu     sync.Mutex
	mu          sync.RWMutex
	closed      bool
}

// NewNacosReader creates a new Nacos reader.
// URI format: nacos://host:port/group/dataId?namespace=public&username=nacos&password=nacos
func NewNacosReader(uri string) (*NacosReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != string(SchemeNacos) {
		return nil, fmt.Errorf("unsupported scheme: %s, expected: %s", u.Scheme, SchemeNacos)
	}

	config, err := parseNacosURI(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Nacos URI: %w", err)
	}

	transport := &http.Transport{}
	if u.Query().Get("tls_insecure") == "true" {
		transport.TLSClientConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}
	}

	closeCtx, closeCancel := context.WithCancel(context.Background())
	n := &NacosReader{
		uri:         uri,
		config:      config,
		closeCtx:    closeCtx,
		closeCancel: closeCancel,
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
	}

	return n, nil
}

// Read reads configuration data from Nacos.
func (n *NacosReader) Read(ctx context.Context) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	ctx, cancel := n.withClose(ctx)
	defer cancel()
	data, err := n.fetch(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(data)) == "" {
		return nil, fmt.Errorf("empty Nacos config %s/%s", n.config.Group, n.config.DataID)
	}

	return data, nil
}

// Subscribe subscribes to Nacos configuration changes.
func (n *NacosReader) Subscribe(ctx context.Context) (<-chan *reader.ReadEvent, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	eventChan := make(chan *reader.ReadEvent, 1)
	ctx, cancel := n.withClose(ctx)
	go func() {
		defer cancel()
		n.subscribe(ctx, eventChan)
	}()
	return eventChan, nil
}

// Close closes the reader and cleans up resources.
func (n *NacosReader) Close() error {
	if n == nil {
		return nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil
	}

	n.closed = true
	n.closeCancel()
	n.client.CloseIdleConnections()
	return nil
}

func (n *NacosReader) withClose(ctx context.Context) (context.Context, context.CancelFunc) {
	requestCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(n.closeCtx, cancel)
	return requestCtx, func() {
		stop()
		cancel()
	}
}

func (n *NacosReader) fetch(ctx context.Context) ([]byte, error) {
	params := n.configValues()
	reqURL := n.config.BaseURL + configPath + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create Nacos config request: %w", err)
	}

	body, err := n.do(req, n.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("get Nacos config %s/%s: %w", n.config.Group, n.config.DataID, err)
	}

	return body, nil
}

func (n *NacosReader) subscribe(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	defer close(eventChan)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		changed, err := n.listen(ctx)
		if err != nil {
			confEvent := reader.NewReadEvent(n.uri, nil, err)
			select {
			case eventChan <- confEvent:
			case <-ctx.Done():
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(n.config.RetryDelay):
				continue
			}
		}

		if changed {
			data, err := n.fetch(ctx)
			confEvent := reader.NewReadEvent(n.uri, data, err)
			select {
			case eventChan <- confEvent:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (n *NacosReader) listen(ctx context.Context) (bool, error) {
	params := url.Values{}
	params.Set(listeningConfigsParamName, n.listenPayload())

	timeout := n.config.ListenTimeout + time.Second
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, n.config.BaseURL+configListenPath, strings.NewReader(params.Encode()))
	if err != nil {
		return false, fmt.Errorf("create Nacos listener request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("Long-Pulling-Timeout", strconv.FormatInt(n.config.ListenTimeout.Milliseconds(), 10))

	body, err := n.do(req, timeout)
	if err != nil {
		return false, fmt.Errorf("listen Nacos config %s/%s: %w", n.config.Group, n.config.DataID, err)
	}

	return strings.TrimSpace(string(body)) != "", nil
}

func (n *NacosReader) do(req *http.Request, timeout time.Duration) ([]byte, error) {
	if err := n.ensureToken(req.Context()); err != nil {
		return nil, err
	}

	values := req.URL.Query()
	if n.accessToken != "" {
		values.Set("accessToken", n.accessToken)
		req.URL.RawQuery = values.Encode()
	}

	client := *n.client
	client.Timeout = timeout
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (n *NacosReader) ensureToken(ctx context.Context) error {
	n.tokenMu.Lock()
	defer n.tokenMu.Unlock()

	if n.config.Username == "" || n.accessToken != "" {
		return nil
	}

	params := url.Values{}
	params.Set("username", n.config.Username)
	params.Set("password", n.config.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.config.BaseURL+authLoginPath, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("create Nacos auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("login Nacos: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read auth response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login Nacos failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode Nacos auth response: %w", err)
	}
	if result.AccessToken == "" {
		return fmt.Errorf("empty Nacos access token")
	}

	n.accessToken = result.AccessToken
	return nil
}

func (n *NacosReader) configValues() url.Values {
	values := url.Values{}
	values.Set("dataId", n.config.DataID)
	values.Set("group", n.config.Group)
	if n.config.Namespace != "" {
		values.Set("tenant", n.config.Namespace)
	}
	return values
}

func (n *NacosReader) listenPayload() string {
	tenant := n.config.Namespace
	return n.config.DataID + splitConfigInner +
		n.config.Group + splitConfigInner +
		"" + splitConfigInner +
		tenant + splitConfigInner +
		splitConfig
}

func parseNacosURI(u *url.URL) (*NacosConfig, error) {
	if strings.TrimSpace(u.Hostname()) == "" {
		return nil, fmt.Errorf("host is required")
	}

	group, dataID, err := parsePath(u.EscapedPath())
	if err != nil {
		return nil, err
	}

	port := DefaultPort
	if u.Port() != "" {
		parsedPort, err := strconv.ParseUint(u.Port(), 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
		port = int(parsedPort)
	}

	query := u.Query()
	timeout, err := parseDuration(query, "timeout", DefaultTimeout)
	if err != nil {
		return nil, err
	}
	listenTimeout, err := parseDuration(query, "listen_timeout", DefaultListenTimeout)
	if err != nil {
		return nil, err
	}
	retryDelay, err := parseDuration(query, "retry_delay", DefaultRetryDelay)
	if err != nil {
		return nil, err
	}

	serverScheme := firstNonEmpty(query.Get("server_scheme"), "http")
	contextPath := firstNonEmpty(query.Get("context_path"), DefaultContextPath)
	if !strings.HasPrefix(contextPath, "/") {
		contextPath = "/" + contextPath
	}
	contextPath = strings.TrimSuffix(contextPath, "/")

	namespace := query.Get("namespace")
	if namespace == "public" {
		namespace = ""
	}

	username := query.Get("username")
	password := query.Get("password")
	if u.User != nil {
		if parsedUsername := u.User.Username(); parsedUsername != "" {
			username = parsedUsername
		}
		if parsedPassword, ok := u.User.Password(); ok {
			password = parsedPassword
		}
	}

	baseURL := fmt.Sprintf("%s://%s:%d%s", serverScheme, u.Hostname(), port, contextPath)

	return &NacosConfig{
		BaseURL:       baseURL,
		Group:         group,
		DataID:        dataID,
		Namespace:     namespace,
		Username:      username,
		Password:      password,
		Timeout:       timeout,
		ListenTimeout: listenTimeout,
		RetryDelay:    retryDelay,
	}, nil
}

func parsePath(path string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid Nacos URI path, expected: nacos://host:port/{group}/{dataId}")
	}

	group, err := url.PathUnescape(parts[0])
	if err != nil {
		return "", "", fmt.Errorf("invalid group: %w", err)
	}
	dataID, err := url.PathUnescape(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("invalid dataId: %w", err)
	}
	if strings.TrimSpace(group) == "" {
		return "", "", fmt.Errorf("group is required")
	}
	if strings.TrimSpace(dataID) == "" {
		return "", "", fmt.Errorf("dataId is required")
	}

	return group, dataID, nil
}

func parseDuration(query url.Values, key string, defaultValue time.Duration) (time.Duration, error) {
	value := query.Get(key)
	if value == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s format: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return duration, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
