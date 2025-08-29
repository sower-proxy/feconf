package xml

import (
	stdxml "encoding/xml"
	"testing"

	"github.com/sower-proxy/conf/decoder"
)

func TestXMLDecoder_Decode(t *testing.T) {
	decoder := NewXMLDecoder()

	t.Run("valid XML with simple structure", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <name>test</name>
    <value>123</value>
    <enabled>true</enabled>
</config>`)

		type Config struct {
			XMLName stdxml.Name `xml:"config"`
			Name    string      `xml:"name"`
			Value   int         `xml:"value"`
			Enabled bool        `xml:"enabled"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Name != "test" {
			t.Errorf("expected name 'test', got: %s", result.Name)
		}
		if result.Value != 123 {
			t.Errorf("expected value 123, got: %d", result.Value)
		}
		if result.Enabled != true {
			t.Errorf("expected enabled true, got: %v", result.Enabled)
		}
	})

	t.Run("valid XML with nested elements", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <title>XML Example</title>
    <database>
        <host>localhost</host>
        <port>5432</port>
        <enabled>true</enabled>
    </database>
    <server>
        <host>127.0.0.1</host>
        <port>8080</port>
    </server>
</config>`)

		type Database struct {
			Host    string `xml:"host"`
			Port    int    `xml:"port"`
			Enabled bool   `xml:"enabled"`
		}

		type Server struct {
			Host string `xml:"host"`
			Port int    `xml:"port"`
		}

		type Config struct {
			XMLName  stdxml.Name `xml:"config"`
			Title    string      `xml:"title"`
			Database Database    `xml:"database"`
			Server   Server      `xml:"server"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Title != "XML Example" {
			t.Errorf("expected title 'XML Example', got: %s", result.Title)
		}
		if result.Database.Host != "localhost" {
			t.Errorf("expected database host 'localhost', got: %s", result.Database.Host)
		}
		if result.Database.Port != 5432 {
			t.Errorf("expected database port 5432, got: %d", result.Database.Port)
		}
		if result.Database.Enabled != true {
			t.Errorf("expected database enabled true, got: %v", result.Database.Enabled)
		}
		if result.Server.Host != "127.0.0.1" {
			t.Errorf("expected server host '127.0.0.1', got: %s", result.Server.Host)
		}
		if result.Server.Port != 8080 {
			t.Errorf("expected server port 8080, got: %d", result.Server.Port)
		}
	})

	t.Run("valid XML with arrays", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <tags>
        <tag>web</tag>
        <tag>api</tag>
        <tag>service</tag>
    </tags>
</config>`)

		type Config struct {
			XMLName stdxml.Name `xml:"config"`
			Tags    struct {
				Items []string `xml:"tag"`
			} `xml:"tags"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(result.Tags.Items) != 3 {
			t.Errorf("expected 3 tags, got: %d", len(result.Tags.Items))
		}
		if result.Tags.Items[0] != "web" {
			t.Errorf("expected first tag 'web', got: %s", result.Tags.Items[0])
		}
		if result.Tags.Items[1] != "api" {
			t.Errorf("expected second tag 'api', got: %s", result.Tags.Items[1])
		}
		if result.Tags.Items[2] != "service" {
			t.Errorf("expected third tag 'service', got: %s", result.Tags.Items[2])
		}
	})

	t.Run("valid XML with attributes", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <server id="main" type="http">
        <name>web-server</name>
        <port>8080</port>
    </server>
</config>`)

		type Server struct {
			ID   string `xml:"id,attr"`
			Type string `xml:"type,attr"`
			Name string `xml:"name"`
			Port int    `xml:"port"`
		}

		type Config struct {
			XMLName stdxml.Name `xml:"config"`
			Server  Server      `xml:"server"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Server.ID != "main" {
			t.Errorf("expected server id 'main', got: %s", result.Server.ID)
		}
		if result.Server.Type != "http" {
			t.Errorf("expected server type 'http', got: %s", result.Server.Type)
		}
		if result.Server.Name != "web-server" {
			t.Errorf("expected server name 'web-server', got: %s", result.Server.Name)
		}
		if result.Server.Port != 8080 {
			t.Errorf("expected server port 8080, got: %d", result.Server.Port)
		}
	})

	t.Run("valid XML with repeated elements", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <services>
        <service>
            <name>web</name>
            <port>8080</port>
        </service>
        <service>
            <name>api</name>
            <port>9000</port>
        </service>
    </services>
</config>`)

		type Service struct {
			Name string `xml:"name"`
			Port int    `xml:"port"`
		}

		type Config struct {
			XMLName  stdxml.Name `xml:"config"`
			Services struct {
				Items []Service `xml:"service"`
			} `xml:"services"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(result.Services.Items) != 2 {
			t.Errorf("expected 2 services, got: %d", len(result.Services.Items))
		}
		if result.Services.Items[0].Name != "web" {
			t.Errorf("expected first service name 'web', got: %s", result.Services.Items[0].Name)
		}
		if result.Services.Items[0].Port != 8080 {
			t.Errorf("expected first service port 8080, got: %d", result.Services.Items[0].Port)
		}
		if result.Services.Items[1].Name != "api" {
			t.Errorf("expected second service name 'api', got: %s", result.Services.Items[1].Name)
		}
		if result.Services.Items[1].Port != 9000 {
			t.Errorf("expected second service port 9000, got: %d", result.Services.Items[1].Port)
		}
	})

	t.Run("valid XML with CDATA", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <name>test</name>
    <description><![CDATA[This is a test configuration with <special> characters & symbols]]></description>
</config>`)

		type Config struct {
			XMLName     stdxml.Name `xml:"config"`
			Name        string      `xml:"name"`
			Description string      `xml:"description"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Name != "test" {
			t.Errorf("expected name 'test', got: %s", result.Name)
		}
		expected := "This is a test configuration with <special> characters & symbols"
		if result.Description != expected {
			t.Errorf("expected description '%s', got: %s", expected, result.Description)
		}
	})

	t.Run("XML without declaration", func(t *testing.T) {
		data := []byte(`<config>
    <name>test</name>
    <value>456</value>
</config>`)

		type Config struct {
			XMLName stdxml.Name `xml:"config"`
			Name    string      `xml:"name"`
			Value   int         `xml:"value"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Name != "test" {
			t.Errorf("expected name 'test', got: %s", result.Name)
		}
		if result.Value != 456 {
			t.Errorf("expected value 456, got: %d", result.Value)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		data := []byte("")
		type Config struct {
			Name string `xml:"name"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for empty data")
		}
	})

	t.Run("nil target", func(t *testing.T) {
		data := []byte(`<config><name>test</name></config>`)

		err := decoder.Unmarshal(data, nil)
		if err == nil {
			t.Fatal("expected error for nil target")
		}
	})

	t.Run("invalid XML", func(t *testing.T) {
		data := []byte(`<config>
    <name>test</name
    <value>123</value>
</config>`)
		type Config struct {
			Name  string `xml:"name"`
			Value int    `xml:"value"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for invalid XML")
		}
	})

	t.Run("XML with mixed content", func(t *testing.T) {
		data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <message>Hello <emphasis>World</emphasis>!</message>
</config>`)

		type Config struct {
			XMLName stdxml.Name `xml:"config"`
			Message string      `xml:"message"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// XML unmarshaling with mixed content typically extracts text content
		// In this case, "Hello " + "!" (text around the emphasis tag)
		expected := "Hello !"
		if result.Message != expected {
			t.Errorf("expected message '%s', got: %s", expected, result.Message)
		}
	})
}

func TestXMLDecoder_Registration(t *testing.T) {
	t.Run("format registration", func(t *testing.T) {
		dec, err := decoder.GetDecoder(FormatXML)
		if err != nil {
			t.Fatalf("failed to get XML decoder: %v", err)
		}
		if dec == nil {
			t.Fatal("decoder is nil")
		}
	})

	t.Run("extension mapping", func(t *testing.T) {
		format, err := decoder.FormatFromExtension(".xml")
		if err != nil {
			t.Fatalf("failed to get format from extension: %v", err)
		}
		if format != FormatXML {
			t.Errorf("expected format %s, got: %s", FormatXML, format)
		}
	})

	t.Run("MIME type mapping", func(t *testing.T) {
		testCases := []string{"application/xml", "text/xml"}
		for _, mime := range testCases {
			format, err := decoder.FormatFromMIME(mime)
			if err != nil {
				t.Fatalf("failed to get format from MIME %s: %v", mime, err)
			}
			if format != FormatXML {
				t.Errorf("expected format %s for MIME %s, got: %s", FormatXML, mime, format)
			}
		}
	})
}
