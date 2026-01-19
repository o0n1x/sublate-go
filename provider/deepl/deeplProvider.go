package deepl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	format "github.com/o0n1x/mass-translate-package/format"
	lang "github.com/o0n1x/mass-translate-package/lang"
	provider "github.com/o0n1x/mass-translate-package/provider"
)

func init() {
	provider.Register(provider.DeepL, func(apiKey string) provider.Client {
		return GetDeeplClient(apiKey)
	})
}

var SupportedFromLang = map[lang.Language]bool{
	lang.AutoDetect:      true,
	lang.Arabic:          true,
	lang.Bulgarian:       true,
	lang.Czech:           true,
	lang.Danish:          true,
	lang.German:          true,
	lang.Greek:           true,
	lang.English:         true,
	lang.Spanish:         true,
	lang.Estonian:        true,
	lang.Finnish:         true,
	lang.French:          true,
	lang.Hungarian:       true,
	lang.Indonesian:      true,
	lang.Italian:         true,
	lang.Japanese:        true,
	lang.Korean:          true,
	lang.Lithuanian:      true,
	lang.Latvian:         true,
	lang.NorwegianBokmal: true,
	lang.Dutch:           true,
	lang.Polish:          true,
	lang.Portuguese:      true,
	lang.Romanian:        true,
	lang.Russian:         true,
	lang.Slovak:          true,
	lang.Slovenian:       true,
	lang.Swedish:         true,
	lang.Turkish:         true,
	lang.Ukrainian:       true,
	lang.Chinese:         true,
}

var SupportedToLang = map[lang.Language]bool{
	lang.Arabic:             true,
	lang.Bulgarian:          true,
	lang.Czech:              true,
	lang.Danish:             true,
	lang.German:             true,
	lang.Greek:              true,
	lang.EnglishUS:          true,
	lang.EnglishUK:          true,
	lang.Spanish:            true,
	lang.Estonian:           true,
	lang.Finnish:            true,
	lang.French:             true,
	lang.Hungarian:          true,
	lang.Indonesian:         true,
	lang.Italian:            true,
	lang.Japanese:           true,
	lang.Korean:             true,
	lang.Lithuanian:         true,
	lang.Latvian:            true,
	lang.NorwegianBokmal:    true,
	lang.Dutch:              true,
	lang.Polish:             true,
	lang.PortugueseBrazil:   true,
	lang.PortuguesePortugal: true,
	lang.Romanian:           true,
	lang.Russian:            true,
	lang.Slovak:             true,
	lang.Slovenian:          true,
	lang.Swedish:            true,
	lang.Turkish:            true,
	lang.Ukrainian:          true,
	lang.ChineseSimplified:  true,
	lang.ChineseTraditional: true,
}

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

type Translations struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

type DeepLClient struct {
	Client  *http.Client
	BaseURL *url.URL
	APIKey  string
	IsFree  bool
}

const (
	ProAPIHost             = "https://api.deepl.com"
	FreeAPIHost            = "https://api-free.deepl.com"
	APIVersion             = "v2"
	accountPlanIdentifyKey = ":fx"
)

func isFreeAccount(authKey string) bool {
	return strings.HasSuffix(authKey, accountPlanIdentifyKey)
}

func apiHost(authKey string) string {
	if isFreeAccount(authKey) {
		return FreeAPIHost
	}
	return ProAPIHost
}

func GetDeeplClient(apiKey string) *DeepLClient {
	u, _ := url.Parse(apiHost(apiKey))
	baseURL := u.JoinPath(APIVersion)
	return &DeepLClient{
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: baseURL,
		APIKey:  apiKey,
		IsFree:  isFreeAccount(apiKey),
	}
}

// will verify the input like from/to lang is valid and use the appropriate helper function to get translation
func (c *DeepLClient) Translate(ctx context.Context, req provider.Request) (provider.Response, error) {
	//validate lang
	if !SupportedFromLang[req.From] {
		return provider.Response{}, fmt.Errorf("Error from DeepLProvider: Invalid Source Language : %v", req.From)
	}
	if !SupportedToLang[req.To] {
		return provider.Response{}, fmt.Errorf("Error from DeepLProvider: Invalid Target Language : %v", req.To)
	}

	//validate type
	if !SupportedFormats[req.ReqType] {
		return provider.Response{}, fmt.Errorf("Error from DeepL Provider: Invalid Request Type : %v", req.ReqType.String())
	}

	if len(req.Text) == 0 {
		return provider.Response{}, fmt.Errorf("Error from DeepL Provider: Invalid Request no FilePath in Request.text given")
	}

	if req.ReqType == format.Text {
		return c.translateText(ctx, req.Text, req.From, req.To)
	}
	return c.translateDoc(ctx, req.Text[0], req.From, req.To)
}

// will approx get the cost without an api call
// TODO: calculate the cost of specific types files too like srt
func (c *DeepLClient) GetCost(req provider.Request) float32 {
	switch req.ReqType {
	case format.Text:
		totalChars := 0

		for _, s := range req.Text {
			// Count the number of characters (runes) in the current string
			count := utf8.RuneCountInString(s)
			totalChars += count
		}
		return (25.0 * float32(totalChars)) / 1000000 // https://www.deepl.com/en/pro#developer
	default:
		return -1
	}

}

func (c *DeepLClient) Name() provider.Provider {
	return provider.DeepL
}

func (c *DeepLClient) Version() string {
	return APIVersion
}

func (c *DeepLClient) translateText(ctx context.Context, text []string, from lang.Language, to lang.Language) (provider.Response, error) {

	params := struct {
		Text       []string `json:"text"`
		TargetLang string   `json:"target_lang"`
		SourceLang string   `json:"source_lang,omitempty"`
	}{
		Text:       text,
		TargetLang: to.String(),
	}

	if from != lang.AutoDetect {
		params.SourceLang = from.String()
	}

	reqBody, err := json.Marshal(params)
	if err != nil {
		return provider.Response{}, err
	}

	url := c.BaseURL.JoinPath("/translate")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return provider.Response{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.Do(req)
	if err != nil {
		return provider.Response{}, err
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		return provider.Response{}, fmt.Errorf("Text Translation Error: HTTP Error %v \n", res.StatusCode)
	}

	translations := new(Translations)

	err = json.NewDecoder(res.Body).Decode(translations)
	if err != nil {
		return provider.Response{}, err
	}

	if len(translations.Translations) < 1 {
		return provider.Response{}, fmt.Errorf("Text Translation Error: Empty translation array response")
	}

	var textlist []string

	for _, trans := range translations.Translations {
		textlist = append(textlist, trans.Text)
	}

	return provider.Response{
		ResType: provider.Sync,
		Text:    textlist,
	}, nil

}

func (c *DeepLClient) translateDoc(ctx context.Context, docPath string, from lang.Language, to lang.Language) (provider.Response, error) {
	panic("not implemented")
}

func (c *DeepLClient) CheckStatus(ctx context.Context, obj provider.Response) (Status, error) {
	panic("not implemented")
} // expected for obj to contain docid and dockey

func (c *DeepLClient) GetResult(ctx context.Context, obj provider.Response) (Status, error) {
	panic("not implemented")
} // expected for obj to contain docid and dockey
