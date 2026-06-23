# Architecture

## Overview

`feconf` is a Go configuration loading library built around a URI-first contract.
It reads configuration bytes from different backends, decodes them into generic
maps, and finally maps them into typed Go structs with a configurable
`mapstructure` pipeline.

## System Boundary

The library is responsible for:

- locating and reading configuration from supported backends
- decoding raw bytes from supported formats into `map[string]any`
- mapping decoded data into typed Go structs
- providing hook-based value normalization during struct decoding
- exposing one-shot load and subscription-based update flows

The library is not responsible for:

- validating business semantics of application-specific configuration
- managing deployment secrets or external access control policy
- generating configuration templates for downstream applications

## Layers

### Reader Layer

Location: `reader/`

Responsibilities:

- parse reader URIs
- fetch raw configuration bytes from file, HTTP, Redis, Nacos, optional Kubernetes, and other backends
- expose subscription/update capabilities when the backend supports them

### Decoder Layer

Location: `decoder/`

Responsibilities:

- infer or resolve source format
- decode raw bytes into generic structures
- keep format-specific logic isolated per package

### Mapping Layer

Location: `mapstructure.go`, `flag.go`

Responsibilities:

- define the default `mapstructure.DecoderConfig`
- normalize common value shapes with decode hooks
- merge CLI flag overrides into decoded configuration data

Current default hook chain:

1. default zero-value handling
2. environment variable rendering
3. structured string-to-slice parsing for JSON/YAML flow sequences
4. string/number to bool conversion
5. string/number to `slog.Level` conversion
6. string to `time.Duration`
7. CSV string to slice fallback
8. string to basic Go types

### Orchestration Layer

Location: `conf.go`

Responsibilities:

- glue readers, decoders, and mapping together
- provide `Parse` / `ParseCtx` for one-shot loads
- provide `Subscribe` / `SubscribeCtx` for update streams

## Key Data Flow

### Parse Flow

1. Resolve URI and initialize the matching reader
2. Read raw configuration bytes
3. Select the decoder from file extension or content type
4. Decode raw bytes into `map[string]any`
5. Merge flag overrides
6. Apply `mapstructure` hooks and map into the target struct

### Subscription Flow

1. Execute an initial parse
2. Subscribe to reader events
3. Re-decode each valid update payload
4. Emit typed config events to the caller

## Design Decisions

- URI-first configuration source selection keeps backend choice outside the core API
- decoder packages stay format-specific, while value normalization is centralized in the mapping layer
- environment rendering is done before type coercion so downstream hooks see final strings
- string-to-slice parsing accepts structured JSON/YAML literals before CSV fallback in order to support environment-driven list injection without breaking existing comma-separated inputs
- Kubernetes reader lives in the optional `github.com/sower-proxy/feconf/reader/k8s`
  submodule. The root module must not require Kubernetes SDK packages; callers
  only pay that dependency cost when they explicitly import the k8s reader.
- The Kubernetes submodule must be released with the same tag as the root module
  when both change together. Its local `replace github.com/sower-proxy/feconf => ../..`
  is only for repository-local development and is not visible to downstream users.

## Related Docs

- [README.md](README.md)
- `examples/*/README.md`
