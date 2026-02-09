# Mass-translate package

A Go Package for translating Text and Documents using multiple translation providers.

## Features

- Multi-Provider Support: Extensible interface for translation providers
    - Implementing a new Provider simply needs to implement the Client interface
- Type-Safe: Strongly Typed languages and formats
- Document Support: Support Translating PDF,SRT,TXT Documents. (Extendendable)


## Installation

```bash
go get github.com/o0n1x/mass-translate-package
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/o0n1x/mass-translate-package/format"
	"github.com/o0n1x/mass-translate-package/lang"
	"github.com/o0n1x/mass-translate-package/provider"
	mt "github.com/o0n1x/mass-translate-package/translator"
)

func main() {
	// Get a translation client

	deeplAPI := "DEEPL_API_KEY"

	client, err := provider.GetClient(provider.DeepL, deeplAPI)
	if err != nil {
		panic(err)
	}

	// Create a request
	req := provider.Request{
		ReqType: format.Text,
		Text:    []string{"Hello World!"},
		To:      lang.German,
	}

	// Translate
	resp, err := mt.Translate(context.Background(), req, client)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Text) // [Hallo Welt!]
}

```

## Architecture

| Component | Description |
|-----------|-------------|
| 🔴 Main Package | Entry point Package  |
| 🟡 Core Types | Core Packages / Implementation |
| 🔵 Language/Format | Supporting Packages / Types definitions|


![Package Diagram](/diagrams/class%20diagram_package.svg)

## API Reference 

### Core Functions

__Translate a single request__

```go
func Translate(ctx context.Context, req Request, client Client) (Response, error)
```


__Translate multiple requests__
```go
func BatchTranslate(ctx context.Context, reqs []Request, client Client) ([]Response, error)
```
__Get a translation client by provider__
```go
func GetClient(provider Provider, APIKey string) (Client, error)
```

### Client Interface

Implement this to add a new provider:

```go
type Client interface {
	Translate(context.Context, Request) (Response, error) // Translates the Request and returns the response
	GetCost(Request) float32 //calculates cost from request
	GetCharCount(Request) int // counts the total char count from request
	Name() Provider //returns the name of the provider. usually returns proivder const from provider package
	Version() string //returns version of the provider API that is used
}
```


### Supported Providers

|Provider | Status | Documents |
|-----------|-------------|-------------|
|Deepl | ✅ | txt, pdf, srt|


### Types

| Type | Format | example |
|-|-|-|
| Language | ISO 639-1 | EN , DE , JA
| Language (Regional) | ISO 639-1 + ISO 3166-1 | EN-US , PT-BR
| Format | MIME Types / RFC 6838 | text/plain , multipart/form-data


## Related Projects
[Mass-translate Server](https://github.com/o0n1x/mass-translate-server) : REST API server using this package


## Acknowledgments

- [cluttrdev/deepl-go](https://github.com/cluttrdev/deepl-go) - DeepL Go client library, referenced for DeepL API implementation