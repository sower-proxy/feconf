package conf

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sower-proxy/conf/decoder"
	"github.com/sower-proxy/conf/reader"
)

type ConfOpt[T any] struct {
	uri        string
	ParserConf mapstructure.DecoderConfig
	parsedURL  *url.URL
	reader     reader.ConfReader
	decoder    decoder.ConfDecoder
	rawData    []byte
	parsedData map[string]any
}

func New[T any](uri string) *ConfOpt[T] {
	return &ConfOpt[T]{
		uri:        uri,
		ParserConf: DefaultParserConfig,
	}
}

func (c *ConfOpt[T]) parseUri() (err error) {
	c.parsedURL, err = reader.ParseURI(c.uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}

	c.reader, err = reader.NewReader(c.parsedURL.String())
	if err != nil {
		return fmt.Errorf("failed to get reader for URI: %w", err)
	}

	format, err := c.getFormat()
	if err != nil {
		return fmt.Errorf("failed to determine format: %w", err)
	}

	c.decoder, err = decoder.GetDecoder(format)
	if err != nil {
		return fmt.Errorf("failed to get decoder for format %s: %w", format, err)
	}

	return nil
}
func (c *ConfOpt[T]) getFormat() (decoder.Format, error) {
	if c.parsedURL == nil {
		return "", fmt.Errorf("URI not parsed")
	}

	ext := filepath.Ext(c.parsedURL.Path)
	if ext != "" {
		return decoder.FormatFromExtension(ext)
	}

	contentType := c.parsedURL.Query().Get("content-type")
	if contentType != "" {
		return decoder.FormatFromMIME(contentType)
	}

	return "", fmt.Errorf("cannot determine format from URI: %s", c.uri)
}

func (c *ConfOpt[T]) readData(ctx context.Context) error {
	if c.reader == nil {
		return fmt.Errorf("reader not initialized")
	}

	data, err := c.reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("empty configuration data")
	}

	c.rawData = data
	return nil
}

func (c *ConfOpt[T]) decode() error {
	if len(c.rawData) == 0 {
		return fmt.Errorf("no data to decode")
	}

	if c.decoder == nil {
		return fmt.Errorf("decoder not initialized")
	}

	var parsedData map[string]any
	if err := c.decoder.Decode(c.rawData, &parsedData); err != nil {
		return fmt.Errorf("failed to decode configuration data: %w", err)
	}

	c.parsedData = parsedData
	return nil
}

func (c *ConfOpt[T]) Load() (*T, error) { return c.LoadCtx(context.Background()) }
func (c *ConfOpt[T]) LoadCtx(ctx context.Context) (*T, error) {
	if err := c.parseUri(); err != nil {
		return nil, fmt.Errorf("failed to parse URI: %w", err)
	}

	if err := c.readData(ctx); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	if err := c.decode(); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	var result T
	c.ParserConf.Result = &result
	mapDecoder, err := mapstructure.NewDecoder(&c.ParserConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}

	if err := mapDecoder.Decode(c.parsedData); err != nil {
		return nil, fmt.Errorf("failed to decode to struct: %w", err)
	}

	return &result, nil
}

type ConfEvent[T any] struct {
	SourceURI string    `json:"source_uri"`
	Timestamp time.Time `json:"timestamp"`
	Error     error     `json:"error,omitempty"`
	Config    *T        `json:"config,omitempty"`
}

func (c *ConfEvent[T]) IsValid() bool {
	return c != nil && c.Error == nil && c.Config != nil
}

func (c *ConfOpt[T]) Subscribe() (<-chan *ConfEvent[T], error) {
	return c.SubscribeCtx(context.Background())
}

func (c *ConfOpt[T]) SubscribeCtx(ctx context.Context) (<-chan *ConfEvent[T], error) {
	initialResult, err := c.Load()
	if err != nil {
		return nil, err
	}

	eventChan, err := c.reader.Subscribe(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to configuration changes: %w", err)
	}

	confEventChan := make(chan *ConfEvent[T], 1)

	// Send initial configuration event
	initialEvent := &ConfEvent[T]{
		SourceURI: c.uri,
		Timestamp: time.Now(),
		Error:     nil,
		Config:    initialResult,
	}
	confEventChan <- initialEvent

	go func() {
		defer close(confEventChan)

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-eventChan:
				if !ok {
					return
				}

				confEvent := &ConfEvent[T]{
					SourceURI: event.SourceURI,
					Timestamp: event.Timestamp,
					Error:     event.Error,
				}

				if event.IsValid() {
					c.rawData = event.Data
					if err := c.decode(); err != nil {
						confEvent.Error = err
						confEventChan <- confEvent
						continue
					}

					var result T
					c.ParserConf.Result = &result
					mapDecoder, err := mapstructure.NewDecoder(&c.ParserConf)
					if err != nil {
						confEvent.Error = err
						confEventChan <- confEvent
						continue
					}

					if err := mapDecoder.Decode(c.parsedData); err != nil {
						confEvent.Error = err
						confEventChan <- confEvent
						continue
					}

					confEvent.Config = &result
					confEvent.Error = nil
				}

				confEventChan <- confEvent
			}
		}
	}()

	return confEventChan, nil
}

func (c *ConfOpt[T]) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
