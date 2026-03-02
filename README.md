# Sublate-go

A Go Package for translating Text and Documents using multiple translation providers.

## Features

- Multi-Provider Support: Extensible interface for translation providers
    - Implementing a new Provider simply needs to implement the SyncClient/AsyncClient interfaces
	- Translation Package works with any Client that implements the interfaces
- Type-Safe: Strongly Typed languages, formats, and errors
- Document Support: Support Translating PDF,SRT,TXT Documents (Extendable)


## Installation

```bash
go get github.com/o0n1x/sublate-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/o0n1x/sublate-go/format"
	"github.com/o0n1x/sublate-go/lang"
	"github.com/o0n1x/sublate-go/provider"
	_ "github.com/o0n1x/sublate-go/provider/deepl" // important! 
	sublate "github.com/o0n1x/sublate-go/translator"
	
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
	resp, err := sublate.Translate(context.Background(), req, client)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Text) // [Hallo Welt!]
}

```
> [!IMPORTANT]
> Provider packages use blank imports (`_`) to self-register via `init()`. This is the same pattern used by Go's `database/sql` package. Without this import, `provider.GetClient()` won't know about the provider.


## Architecture


| Component | Description |
|-----------|-------------|
| 🔴 Translator Package | Entry point  |
| 🟡 Core Types | Interfaces, structs, and provider registry |
| 🔵 Provider Implementation | Concrete provider packages (e.g. DeepL) |
| ⚫ Supporting Types | Enumerations, error types, and shared definitions (Language, Format, ErrorCode) |



![Class Diagram](/diagrams/class%20diagram_package.svg)

## API Reference 

### Core Functions

__Translate a single request__

```go
func Translate(ctx context.Context, req Request, client Client) (Response, error)
```


__Translate multiple requests__
```go
func BatchTranslate(ctx context.Context, reqs []Request, client Client) ([]Response, []error)
```
__Get a translation client by provider__
```go
func GetClient(provider Provider, APIKey string) (Client, error)
```

### Client Interface

All Clients implements the generalized client interface:
```go
type Client interface {
	GetCost(Request) float32 //calculates cost from request
	GetCharCount(Request) int // counts the total char count from request
	Name() Provider //returns the name of the provider. usually returns provider const from provider package
	Version() string //returns version of the provider API that is used
}
```
> [!NOTE]
> Clients only implementing the generalized Client wont be able to translate anything. it should either implement SyncClient or AsyncClient.


Synchronous clients should implement the SyncClient interface:

```go
type SyncClient interface {
	Translate(context.Context, Request) (Response, error) // Translates the Request and returns a response
	Client
}
```

Asynchronous clients should implement the AsyncClient interface:
```go
type AsyncClient interface {
	AsyncTranslate(context.Context, Request) (AsyncResponse, error) //  Sends a request to the Provider and returns an async response
	CheckStatus(context.Context, AsyncResponse) (JobStatus, error) // Check Status of a Document wit the provider
	GetResult(context.Context, AsyncResponse) (Response, error) // Gets result document from provider
	Client
}
```

> [!NOTE]
> Providers can implement both Sync and AsyncClient interfaces.

### Supported Providers

|Provider | Sync | Async | Supports |
|-----------|-------------|-------------|-------------|
|Deepl | ✅ | ✅ | txt, pdf, srt, string|


### Types

| Type | Format | example |
|-|-|-|
| Language | ISO 639-1 | EN , DE , JA
| Language (Regional) | ISO 639-1 + ISO 3166-1 | EN-US , PT-BR
| Format | MIME Types / RFC 6838 | text/plain , multipart/form-data


## Related Projects
[Sublate](https://github.com/o0n1x/Sublate) : REST API server using this package


## Acknowledgments

- [cluttrdev/deepl-go](https://github.com/cluttrdev/deepl-go) - DeepL Go client library, referenced for DeepL API implementation