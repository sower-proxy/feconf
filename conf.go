package feconf

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sower-proxy/feconf/decoder"
	"github.com/sower-proxy/feconf/reader"
)

type ConfOpt[T any] struct {
	uri         string
	originalURI string
	flagName    string
	ParserConf  mapstructure.DecoderConfig
	parsedURL   *url.URL
	reader      reader.ConfReader
	decoder     decoder.ConfDecoder
	rawData     []byte
	parsedData  map[string]any
}

func New[T any](flag string, uris ...string) *ConfOpt[T] {
	var uri string
	var r reader.ConfReader
	for _, u := range uris {
		var err error
		if r, err = reader.NewReader(u); err == nil {
			uri = u
			break
		}
	}

	conf := &ConfOpt[T]{
		uri:         uri,
		originalURI: uri,
		flagName:    flag,
		ParserConf:  DefaultParserConfig,
		reader:      r,
	}
	conf.registerFlags()
	return conf
}

func (c *ConfOpt[T]) parseUri() error {
	var err error
	if c.parsedURL, err = reader.ParseURI(c.uri); err != nil {
		return fmt.Errorf("parse URI: %w", err)
	}

	if c.uri != c.originalURI && c.reader != nil {
		c.reader.Close()
		c.reader = nil
	}

	if c.reader == nil {
		if c.reader, err = reader.NewReader(c.parsedURL.String()); err != nil {
			return fmt.Errorf("create reader: %w", err)
		}
	}

	format, err := c.getFormat()
	if err != nil {
		return fmt.Errorf("determine format: %w", err)
	}
	if c.decoder, err = decoder.GetDecoder(format); err != nil {
		return fmt.Errorf("get decoder for %s: %w", format, err)
	}
	return nil
}

func (c *ConfOpt[T]) getFormat() (decoder.Format, error) {
	if c.parsedURL == nil {
		return "", fmt.Errorf("URI not parsed")
	}
	if ext := filepath.Ext(c.parsedURL.Path); ext != "" {
		return decoder.FormatFromExtension(ext)
	}
	if ct := c.parsedURL.Query().Get("content-type"); ct != "" {
		return decoder.FormatFromMIME(ct)
	}
	return "", fmt.Errorf("cannot determine format from URI: %s", c.uri)
}

func (c *ConfOpt[T]) readData(ctx context.Context) error {
	if c.reader == nil {
		return fmt.Errorf("reader not initialized")
	}
	data, err := c.reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("read configuration: %w", err)
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
	var data map[string]any
	if err := c.decoder.Unmarshal(c.rawData, &data); err != nil {
		return fmt.Errorf("decode configuration: %w", err)
	}
	c.parsedData = data
	return nil
}

func (c *ConfOpt[T]) loadAndDecode(ctx context.Context) error {
	if err := c.parseUri(); err != nil {
		return err
	}
	if err := c.readData(ctx); err != nil {
		return err
	}
	return c.decode()
}

func (c *ConfOpt[T]) decodeToStruct(result *T) error {
	c.ParserConf.Result = result
	dec, err := mapstructure.NewDecoder(&c.ParserConf)
	if err != nil {
		return fmt.Errorf("create decoder: %w", err)
	}
	if err := dec.Decode(c.parsedData); err != nil {
		return fmt.Errorf("decode to struct: %w", err)
	}
	return nil
}

func (c *ConfOpt[T]) Parse() (*T, error) { return c.ParseCtx(context.Background()) }

func (c *ConfOpt[T]) ParseCtx(ctx context.Context) (*T, error) {
	var result T
	if err := c.parseCtx(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ConfOpt[T]) parseCtx(ctx context.Context, result *T) error {
	c.parseFlags()
	c.mergeFlagValues()

	if c.uri == "" {
		if c.parsedData == nil {
			c.parsedData = make(map[string]any)
		}
		return c.decodeToStruct(result)
	}

	if err := c.loadAndDecode(ctx); err != nil {
		return err
	}
	c.mergeFlagValues()
	return c.decodeToStruct(result)
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
	initialResult, err := c.Parse()
	if err != nil {
		return nil, err
	}

	eventChan, err := c.reader.Subscribe(ctx)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	confEventChan := make(chan *ConfEvent[T], 1)
	confEventChan <- &ConfEvent[T]{
		SourceURI: c.uri,
		Timestamp: time.Now(),
		Config:    initialResult,
	}

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
					} else {
						var result T
						if err := c.decodeToStruct(&result); err != nil {
							confEvent.Error = err
						} else {
							confEvent.Config = &result
						}
					}
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

func Load[T any](obj *T, uris ...string) error {
	return LoadCtx(context.Background(), obj, uris...)
}

func LoadCtx[T any](ctx context.Context, obj *T, uris ...string) error {
	return New[T]("", uris...).parseCtx(ctx, obj)
}
