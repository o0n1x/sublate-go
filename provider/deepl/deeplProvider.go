package deepl

import (
	"context"
	"net/http"

	format "github.com/o0n1x/mass-translate-package/format"
	lang "github.com/o0n1x/mass-translate-package/lang"
	provider "github.com/o0n1x/mass-translate-package/provider"
)

var SupportedFromLang = map[lang.Language]bool{}

var SupportedToLang = map[lang.Language]bool{}

var SupportedFormats = map[format.Format]bool{
	format.File: true,
	format.Text: true,
}

type Status struct {
	DocumentID       string `json:"document_id"`
	Status           string `json:"status"`
	SecondsRemaining int    `json:"seconds_remaining"`
	BilledCharacters int    `json:"billed_characters"`
	ErrMessage       string `json:"message"`
}

type DeepLClient struct {
	Client *http.Client
	APIkey string
	isFree bool
}

func GetDeeplClient() (*DeepLClient, error)

func (c *DeepLClient) Translate(ctx context.Context, req provider.Request) (provider.Response, error)

func (c *DeepLClient) GetCost(req provider.Request) int

func (c *DeepLClient) Name() string {
	return "DeepL"
}

func (c *DeepLClient) Version() string {
	return "v2"
}

func (c *DeepLClient) CheckStatus(ctx context.Context, obj provider.Response) (Status, error) // expected for obj to contain docid and dockey in keys

func (c *DeepLClient) GetResult(ctx context.Context, obj provider.Response) (Status, error) // expected for obj to contain docid and dockey in keys

func translateText(ctx context.Context, text string, from lang.Language, to lang.Language) (provider.Response, error)

func translateDoc(ctx context.Context, docPath string, from lang.Language, to lang.Language) (provider.Response, error)
