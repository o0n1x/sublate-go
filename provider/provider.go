package provider

import (
	"context"

	format "github.com/o0n1x/mass-translate-package/format"
	lang "github.com/o0n1x/mass-translate-package/lang"
)

type ResponseType int

const (
	Sync ResponseType = iota
	ASync
)

type Response struct {
	ResType ResponseType
	Text    string // if sync then translation is in text field

	// if async then response will be in later fields

	//DeepL Document translation Fields
	DocumentID  string
	DocumentKey string
}

type Request struct {
	ReqType format.Format
	Text    string // can be the actual text, filepah, etc.. depending on the format field
	From    lang.Language
	To      lang.Language
}

type Provider string

const (
	DeepL Provider = "DeepL"
)

type Client interface {
	Translate(context.Context, Request) (Response, error)
	GetCost(Request) int
	Name() Provider
	Version() string
}

// TODO: complete this function
func GetClient(provider Provider) (Client, error) {
	return nil, nil
}
