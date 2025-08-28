package config

import (
	"os"
	"testing"
)

func TestRenderEnv(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "no env vars",
			input:  "simple string",
			output: "simple string",
		},
		{
			name:   "single env var",
			input:  "host: ${TEST_VAR}",
			output: "host: test_value",
		},
		{
			name:   "multiple env vars",
			input:  "mysql://${DB_HOST}:${DB_PORT}/db",
			output: "mysql://localhost:3306/db",
		},
		{
			name:   "env var at beginning",
			input:  "${DB_HOST}/path",
			output: "localhost/path",
		},
		{
			name:   "env var at end",
			input:  "prefix/${TEST_VAR}",
			output: "prefix/test_value",
		},
		{
			name:   "escaped env var",
			input:  "prefix/$${TEST_VAR}",
			output: "prefix/${TEST_VAR}",
		},
		{
			name:   "non-existent env var",
			input:  "${NON_EXISTENT_VAR}",
			output: "",
		},
		{
			name:   "mixed escaped and non-escaped",
			input:  "$${TEST_VAR} and ${DB_HOST}",
			output: "${TEST_VAR} and localhost",
		},
		{
			name:   "empty string",
			input:  "",
			output: "",
		},
		{
			name:   "only dollar sign",
			input:  "$",
			output: "$",
		},
		{
			name:   "malformed env var",
			input:  "${INCOMPLETE",
			output: "${INCOMPLETE",
		},
		{
			name:   "env var with underscore and numbers",
			input:  "${TEST_VAR}_suffix",
			output: "test_value_suffix",
		},
		{
			name:   "env var with default - var exists",
			input:  "${TEST_VAR:-default_value}",
			output: "test_value",
		},
		{
			name:   "env var with default - var does not exist",
			input:  "${NON_EXISTENT_VAR:-default_value}",
			output: "default_value",
		},
		{
			name:   "env var with empty default",
			input:  "${NON_EXISTENT_VAR:-}",
			output: "",
		},
		{
			name:   "multiple env vars with defaults",
			input:  "${NON_EXISTENT_1:-host}:${NON_EXISTENT_2:-3306}/db",
			output: "host:3306/db",
		},
		{
			name:   "mixed existing and default vars",
			input:  "${DB_HOST}:${NON_EXISTENT_PORT:-3306}",
			output: "localhost:3306",
		},
		{
			name:   "default with special characters",
			input:  "${NON_EXISTENT:-default/path:with-special_chars}",
			output: "default/path:with-special_chars",
		},
		{
			name:   "escaped env var with default syntax",
			input:  "$${TEST_VAR:-default}",
			output: "${TEST_VAR:-default}",
		},
		{
			name:   "complex mixed text with multiple vars",
			input:  "App: ${TEST_VAR} version ${NON_EXISTENT_VERSION:-1.0.0}",
			output: "App: test_value version 1.0.0",
		},
		{
			name:   "connection string format",
			input:  "${DB_HOST}:${DB_PORT}/database_name",
			output: "localhost:3306/database_name",
		},
		{
			name:   "url with default values",
			input:  "https://${DOMAIN:-example.com}/api/v${API_VERSION:-1}/users",
			output: "https://example.com/api/v1/users",
		},
		{
			name:   "file path with defaults",
			input:  "/home/${USER_NAME:-user}/projects/${PROJECT:-myproject}/logs",
			output: "/home/user/projects/myproject/logs",
		},
		{
			name:   "mixed escaped and regular with text",
			input:  "Config: $${ESCAPED_VAR:-should-not-expand} and ${TEST_VAR} work together",
			output: "Config: ${ESCAPED_VAR:-should-not-expand} and test_value work together",
		},
		{
			name:   "multiple defaults in sequence",
			input:  "${VAR1:-val1}-${VAR2:-val2}-${VAR3:-val3}",
			output: "val1-val2-val3",
		},
		{
			name:   "complex sentence with mixed vars",
			input:  "Welcome to ${APP_NAME:-MyApp} running on ${HOST:-localhost}:${PORT:-8080}",
			output: "Welcome to MyApp running on localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderEnv(tt.input)
			if result != tt.output {
				t.Errorf("renderEnv(%q) = %q, want %q", tt.input, result, tt.output)
			}
		})
	}
}

func TestPrevByte(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		idx    int
		output byte
	}{
		{
			name:   "index 0",
			value:  "hello",
			idx:    0,
			output: 0,
		},
		{
			name:   "index 1",
			value:  "hello",
			idx:    1,
			output: 'h',
		},
		{
			name:   "index 3",
			value:  "test string",
			idx:    3,
			output: 's',
		},
		{
			name:   "last index",
			value:  "hello",
			idx:    4,
			output: 'l',
		},
		{
			name:   "empty string index 0",
			value:  "",
			idx:    0,
			output: 0,
		},
		{
			name:   "single character index 1",
			value:  "a",
			idx:    1,
			output: 'a',
		},
		{
			name:   "special characters",
			value:  "${}",
			idx:    2,
			output: '{',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prevByte(tt.value, tt.idx)
			if result != tt.output {
				t.Errorf("prevByte(%q, %d) = %q, want %q", tt.value, tt.idx, result, tt.output)
			}
		})
	}
}
