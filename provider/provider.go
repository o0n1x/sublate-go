package provider

import (
	"context"
	"fmt"

	serr "github.com/o0n1x/sublate-go/errors"
	format "github.com/o0n1x/sublate-go/format"
	lang "github.com/o0n1x/sublate-go/lang"
)

// type ResponseType int

// // with async response this should be safely removed
// const (
// 	Sync ResponseType = iota
// 	ASync
// )

type Response struct {
	// ResType ResponseType
	Text []string // if sync then translation is in text field
	// Binary  []byte   // if file then translation in the binary field
	// // if async then response will be in later fields

	// //DeepL Document translation Fields
	// DocumentID  string
	// DocumentKey string
}

type AsyncResponse struct {
	//DeepL Document translation Fields
	DocumentID  string
	DocumentKey string
}

type JobStatus struct {
	Done             bool
	Failed           bool
	SecondsRemaining int
	Message          string
}

type Request struct {
	ReqType  format.Format
	Text     []string // can be the actual text, json, etc.. depending on the format field
	Binary   []byte   // is used when format is File
	FileName string   // is used when format is File
	From     lang.Language
	To       lang.Language
}

type ClientFactory func(apiKey string) Client

var registry = map[Provider]ClientFactory{}

func Register(name Provider, factory ClientFactory) {
	registry[name] = factory
}

type Provider string

const (
	DeepL Provider = "DeepL"
)

type Client interface {
	Translate(context.Context, Request) (Response, error)
	GetCost(Request) float32
	GetCharCount(Request) int
	Name() Provider
	Version() string
}

type AsyncClient interface {
	AsyncTranslate(context.Context, Request) (AsyncResponse, error)
	CheckStatus(context.Context, AsyncResponse) (JobStatus, error)
	GetResult(context.Context, AsyncResponse) (Response, error)
}

func GetClient(name Provider, apiKey string) (Client, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, serr.New(serr.ErrInvalidProvider, "GetClient", "", fmt.Errorf("%s is not a valid provider", name))
	}
	return factory(apiKey), nil
}
