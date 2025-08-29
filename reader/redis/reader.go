package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sower-proxy/conf/reader"
)

const (
	// SchemeRedis represents Redis URI scheme
	SchemeRedis reader.Scheme = "redis"
	// SchemeRediss represents Redis SSL URI scheme
	SchemeRediss reader.Scheme = "rediss"
)

var (
	// DefaultTimeout for Redis operations
	DefaultTimeout = 30 * time.Second
	// DefaultRetryAttempts for failed operations
	DefaultRetryAttempts = 3
	// DefaultRetryDelay between retry attempts
	DefaultRetryDelay = 1 * time.Second
	// DefaultDB is the default Redis database
	DefaultDB = 0
)

// init registers Redis readers
func init() {
	reader.RegisterReader(SchemeRedis, func(uri string) (reader.ConfReader, error) {
		return NewRedisReader(uri)
	})
	reader.RegisterReader(SchemeRediss, func(uri string) (reader.ConfReader, error) {
		return NewRedisReader(uri)
	})
}

// RedisConfig holds Redis client configuration
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	Key          string
	Timeout      time.Duration
	TLSConfig    *tls.Config
	RetryDelay   time.Duration
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

// RedisReader implements ConfReader for Redis-based configuration
type RedisReader struct {
	uri      string
	client   *redis.Client
	config   *RedisConfig
	mu       sync.RWMutex
	closed   bool
	subsChan chan *redis.PubSub
}

// NewRedisReader creates a new Redis reader
func NewRedisReader(uri string) (*RedisReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != string(SchemeRedis) && u.Scheme != string(SchemeRediss) {
		return nil, fmt.Errorf("unsupported scheme: %s, expected: %s or %s", u.Scheme, SchemeRedis, SchemeRediss)
	}

	config := &RedisConfig{
		Timeout:      DefaultTimeout,
		RetryDelay:   DefaultRetryDelay,
		MaxRetries:   DefaultRetryAttempts,
		DB:           DefaultDB,
		PoolSize:     10,
		MinIdleConns: 1,
	}

	// Parse Redis URI configuration
	if err := parseRedisURI(u, config); err != nil {
		return nil, fmt.Errorf("failed to parse Redis URI: %w", err)
	}

	// Validate key is provided
	if config.Key == "" {
		return nil, fmt.Errorf("Redis key must be specified in path")
	}

	// Create Redis options
	opts := &redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.Timeout,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		TLSConfig:    config.TLSConfig,
	}

	client := redis.NewClient(opts)

	return &RedisReader{
		uri:      uri,
		client:   client,
		config:   config,
		subsChan: make(chan *redis.PubSub, 1),
	}, nil
}

// Read reads configuration data from Redis key
func (r *RedisReader) Read(ctx context.Context) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	return r.fetchWithRetry(ctx)
}

// Subscribe subscribes to Redis keyspace notifications for real-time updates
func (r *RedisReader) Subscribe(ctx context.Context) (<-chan *reader.ReadEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	eventChan := make(chan *reader.ReadEvent, 1)

	// Check if keyspace notifications are enabled
	if err := r.ensureKeyspaceNotifications(ctx); err != nil {
		close(eventChan)
		return nil, fmt.Errorf("failed to enable keyspace notifications: %w", err)
	}

	go r.subscribeKeyspace(ctx, eventChan)

	return eventChan, nil
}

// Close closes the reader and cleans up resources
func (r *RedisReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	// Close any active subscriptions
	select {
	case pubsub := <-r.subsChan:
		if pubsub != nil {
			pubsub.Close()
		}
	default:
	}

	if r.client != nil {
		return r.client.Close()
	}

	return nil
}

// fetchWithRetry performs Redis GET with retry mechanism
func (r *RedisReader) fetchWithRetry(ctx context.Context) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(r.config.RetryDelay):
			}
		}

		data, err := r.fetch(ctx)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", r.config.MaxRetries, lastErr)
}

// fetch performs single Redis GET operation
func (r *RedisReader) fetch(ctx context.Context) ([]byte, error) {
	result := r.client.Get(ctx, r.config.Key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, fmt.Errorf("key '%s' not found", r.config.Key)
		}
		return nil, fmt.Errorf("failed to get key '%s': %w", r.config.Key, result.Err())
	}

	data, err := result.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert result to bytes: %w", err)
	}

	return data, nil
}

// ensureKeyspaceNotifications enables keyspace notifications if not already enabled
func (r *RedisReader) ensureKeyspaceNotifications(ctx context.Context) error {
	// Check current notification settings
	config := r.client.ConfigGet(ctx, "notify-keyspace-events")
	if config.Err() != nil {
		return fmt.Errorf("failed to check keyspace notifications config: %w", config.Err())
	}

	currentConfig := ""
	if len(config.Val()) >= 2 {
		currentConfig = config.Val()[1].(string)
	}

	// Enable keyspace notifications if not already enabled
	// We need 'K' (keyspace events) and either 's' (string commands) or '$' (generic commands)
	needsKeyspace := !strings.Contains(currentConfig, "K")
	needsCommands := !strings.Contains(currentConfig, "s") && !strings.Contains(currentConfig, "$")
	
	if needsKeyspace || needsCommands {
		newConfig := currentConfig
		if needsKeyspace {
			newConfig += "K"
		}
		if needsCommands {
			newConfig += "$" // Use generic commands instead of string commands
		}
		result := r.client.ConfigSet(ctx, "notify-keyspace-events", newConfig)
		if result.Err() != nil {
			return fmt.Errorf("failed to enable keyspace notifications: %w", result.Err())
		}
	}

	return nil
}

// subscribeKeyspace subscribes to Redis keyspace notifications
func (r *RedisReader) subscribeKeyspace(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	defer close(eventChan)

	// Create keyspace pattern for the specific key
	keyspacePattern := fmt.Sprintf("__keyspace@%d__:%s", r.config.DB, r.config.Key)
	
	// Debug: Log subscription details
	fmt.Printf("Redis: Subscribing to pattern: %s\n", keyspacePattern)

	pubsub := r.client.PSubscribe(ctx, keyspacePattern)
	defer pubsub.Close()

	// Store pubsub for cleanup
	select {
	case r.subsChan <- pubsub:
	default:
		// Channel full, proceed anyway
	}

	// Send initial read
	if data, err := r.fetch(ctx); err == nil {
		fmt.Printf("Redis: Sending initial config event\n")
		confEvent := reader.NewReadEvent(r.uri, data, nil)
		select {
		case eventChan <- confEvent:
		case <-ctx.Done():
			return
		}
	}

	// Listen for notifications
	fmt.Printf("Redis: Starting keyspace notification listener\n")
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Redis: Context done, stopping listener\n")
			return
		case msg, ok := <-ch:
			if !ok {
				fmt.Printf("Redis: Channel closed, stopping listener\n")
				return
			}

			// Handle keyspace notification
			if msg != nil {
				fmt.Printf("Redis: Received keyspace notification: %+v\n", msg)
				// Fetch the updated value
				data, err := r.fetch(ctx)
				confEvent := reader.NewReadEvent(r.uri, data, err)
				select {
				case eventChan <- confEvent:
					fmt.Printf("Redis: Sent config update event\n")
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// parseRedisURI parses Redis URI into configuration
func parseRedisURI(u *url.URL, config *RedisConfig) error {
	// Set address
	config.Addr = u.Host
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}

	// Extract password from userinfo
	if u.User != nil {
		config.Password = u.User.Username()
		if password, ok := u.User.Password(); ok && password != "" {
			config.Password = password
		}
	}

	// Extract key from path
	if u.Path != "" {
		config.Key = strings.TrimPrefix(u.Path, "/")
	}

	// Parse query parameters
	query := u.Query()

	// Parse database number
	if dbStr := query.Get("db"); dbStr != "" {
		db, err := strconv.Atoi(dbStr)
		if err != nil {
			return fmt.Errorf("invalid db format: %w", err)
		}
		if db < 0 {
			return fmt.Errorf("db must be non-negative")
		}
		config.DB = db
	}

	// Parse timeout
	if timeoutStr := query.Get("timeout"); timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		config.Timeout = timeout
	}

	// Parse retry delay
	if delayStr := query.Get("retry_delay"); delayStr != "" {
		delay, err := time.ParseDuration(delayStr)
		if err != nil {
			return fmt.Errorf("invalid retry_delay format: %w", err)
		}
		config.RetryDelay = delay
	}

	// Parse max retries
	if retriesStr := query.Get("max_retries"); retriesStr != "" {
		retries, err := strconv.Atoi(retriesStr)
		if err != nil {
			return fmt.Errorf("invalid max_retries format: %w", err)
		}
		if retries < 0 {
			return fmt.Errorf("max_retries must be non-negative")
		}
		config.MaxRetries = retries
	}

	// Parse pool size
	if poolSizeStr := query.Get("pool_size"); poolSizeStr != "" {
		poolSize, err := strconv.Atoi(poolSizeStr)
		if err != nil {
			return fmt.Errorf("invalid pool_size format: %w", err)
		}
		if poolSize <= 0 {
			return fmt.Errorf("pool_size must be positive")
		}
		config.PoolSize = poolSize
	}

	// Parse min idle connections
	if minIdleStr := query.Get("min_idle_conns"); minIdleStr != "" {
		minIdle, err := strconv.Atoi(minIdleStr)
		if err != nil {
			return fmt.Errorf("invalid min_idle_conns format: %w", err)
		}
		if minIdle < 0 {
			return fmt.Errorf("min_idle_conns must be non-negative")
		}
		config.MinIdleConns = minIdle
	}

	// Configure TLS for rediss scheme
	if u.Scheme == string(SchemeRediss) {
		config.TLSConfig = &tls.Config{}

		// Parse TLS insecure option
		if insecureStr := query.Get("tls_insecure"); insecureStr == "true" {
			config.TLSConfig.InsecureSkipVerify = true
		}
	}

	return nil
}
