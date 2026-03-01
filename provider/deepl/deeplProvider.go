package deepl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	serr "github.com/o0n1x/sublate-go/errors"
	format "github.com/o0n1x/sublate-go/format"
	lang "github.com/o0n1x/sublate-go/lang"
	provider "github.com/o0n1x/sublate-go/provider"
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

type Documents struct {
	DocID  string `json:"document_id"`
	DocKey string `json:"document_key"`
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
		Client:  &http.Client{},
		BaseURL: baseURL,
		APIKey:  apiKey,
		IsFree:  isFreeAccount(apiKey),
	}
}

// will verify the input like from/to lang is valid and use the appropriate helper function to get translation
func (c *DeepLClient) Translate(ctx context.Context, req provider.Request) (provider.Response, error) {
	//validate lang
	req, err := validateRequest(req)
	if err != nil {
		return provider.Response{}, err
	}

	if req.ReqType == format.Text {
		return c.translateText(ctx, req.Text, req.From, req.To)
	} else {
		return provider.Response{}, serr.New(serr.ErrInvalidRequest, "Translate", string(provider.DeepL), fmt.Errorf("Invalid Request Type %v", req.ReqType.String()))
	}

}

func (c *DeepLClient) AsyncTranslate(ctx context.Context, req provider.Request) (provider.AsyncResponse, error) {
	req, err := validateRequest(req)
	if err != nil {
		return provider.AsyncResponse{}, err
	}

	if req.ReqType == format.File {
		return c.translateDoc(ctx, req.Binary, req.FileName, req.From, req.To)
	} else {
		return provider.AsyncResponse{}, serr.New(serr.ErrInvalidRequest, "AsyncTranslate", string(provider.DeepL), fmt.Errorf("Invalid Request Type %v", req.ReqType.String()))
	}

}

func validateRequest(req provider.Request) (provider.Request, *serr.TranslateError) {
	if req.From == "" {
		req.From = lang.AutoDetect
	}
	if !SupportedFromLang[req.From] {
		return req, serr.New(serr.ErrInvalidLanguage, "validateRequest", string(provider.DeepL), fmt.Errorf("Invalid Source Language %v", req.From))
	}
	if !SupportedToLang[req.To] {
		return req, serr.New(serr.ErrInvalidLanguage, "validateRequest", string(provider.DeepL), fmt.Errorf("Invalid Target Language %v", req.To))
	}

	//validate type
	if !SupportedFormats[req.ReqType] {
		return req, serr.New(serr.ErrInvalidRequest, "validateRequeste", string(provider.DeepL), fmt.Errorf("Invalid Request Type %v", req.ReqType.String()))
	}

	if len(req.Text) == 0 && len(req.FileName) == 0 {
		return req, serr.New(serr.ErrInvalidRequest, "validateRequest", string(provider.DeepL), fmt.Errorf("no text or filename"))
	}

	return req, nil
}

// will approx get the cost without an api call
// TODO: calculate the cost of specific types files too like srt
func (c *DeepLClient) GetCost(req provider.Request) float32 {
	if c.IsFree {
		return 0
	}

	const pricePerMillionChars = 25.0

	return (pricePerMillionChars * float32(c.GetCharCount(req))) / 1_000_000 // https://www.deepl.com/en/pro#developer

}

func (c *DeepLClient) GetCharCount(req provider.Request) int {
	switch req.ReqType {
	case format.Text:
		totalChars := 0

		for _, s := range req.Text {
			count := utf8.RuneCountInString(s)
			totalChars += count
		}
		return totalChars
	default:
		return 0
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
		return provider.Response{}, serr.New(serr.ErrInvalidRequest, "TranslateText", string(provider.DeepL), fmt.Errorf("Error json marshal: %w", err))
	}

	url := c.BaseURL.JoinPath("/translate")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return provider.Response{}, serr.New(serr.ErrHTTP, "TranslateText", string(provider.DeepL), err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.Do(req)
	if err != nil {
		return provider.Response{}, serr.New(serr.ErrNetwork, "TranslateText", string(provider.DeepL), err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices // deepl /translate endpoint returns 200 or 400+ codes no in between
	if !ok {
		return provider.Response{}, serr.New(serr.ErrHTTP, "TranslateText", string(provider.DeepL), fmt.Errorf("response code %v , Trace ID: %v", res.StatusCode, res.Header.Get("X-Trace-ID")))
	}

	translations := new(Translations)

	err = json.NewDecoder(res.Body).Decode(translations)
	if err != nil {
		return provider.Response{}, serr.New(serr.ErrInvalidResponse, "TranslateText", string(provider.DeepL), fmt.Errorf("Error json decoding: %w", err))
	}

	if len(translations.Translations) < 1 {
		return provider.Response{}, serr.New(serr.ErrEmptyResponse, "TranslateText", string(provider.DeepL), fmt.Errorf("Empty translation array response"))
	}

	var textlist []string

	for _, trans := range translations.Translations {
		textlist = append(textlist, trans.Text)
	}

	return provider.Response{
		Text: textlist,
	}, nil

}

func (c *DeepLClient) translateDoc(ctx context.Context, binary []byte, filename string, from lang.Language, to lang.Language) (provider.AsyncResponse, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return provider.AsyncResponse{}, serr.New(serr.ErrInvalidRequest, "TranslateDocument", string(provider.DeepL), fmt.Errorf("Error writing file: %w", err))
	}
	_, err = io.Copy(part, bytes.NewBuffer(binary))
	if err != nil {
		return provider.AsyncResponse{}, serr.New(serr.ErrIO, "TranslateDocument", string(provider.DeepL), fmt.Errorf("Error copying file: %w", err))
	}

	if from != lang.AutoDetect {
		err = writer.WriteField("source_lang", from.String())
		if err != nil {
			return provider.AsyncResponse{}, serr.New(serr.ErrIO, "TranslateDocument", string(provider.DeepL), fmt.Errorf("Error writing source_lang to request body: %w", err))
		}
	}

	err = writer.WriteField("target_lang", to.String())
	if err != nil {
		return provider.AsyncResponse{}, serr.New(serr.ErrHTTP, "TranslateDocument", string(provider.DeepL), fmt.Errorf("Error writing target_lang to request body: %w", err))
	}

	writer.Close()

	url := c.BaseURL.JoinPath("/document")
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), body)
	if err != nil {
		return provider.AsyncResponse{}, serr.New(serr.ErrHTTP, "TranslateDocument", string(provider.DeepL), fmt.Errorf("Error creating request: %w", err))
	}
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := c.Client.Do(req)
	if err != nil {
		return provider.AsyncResponse{}, serr.New(serr.ErrNetwork, "TranslateDocument", string(provider.DeepL), err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		return provider.AsyncResponse{}, serr.New(serr.ErrHTTP, "TranslateDocument", string(provider.DeepL), fmt.Errorf("response code %v , Trace ID: %v", res.StatusCode, res.Header.Get("X-Trace-ID")))
	}

	document := new(Documents)

	err = json.NewDecoder(res.Body).Decode(document)
	if err != nil {
		return provider.AsyncResponse{}, serr.New(serr.ErrInvalidResponse, "TranslateText", string(provider.DeepL), fmt.Errorf("Error json decoding: %w", err))
	}

	return provider.AsyncResponse{
		DocumentID:  document.DocID,
		DocumentKey: document.DocKey,
	}, nil

}

func (c *DeepLClient) CheckStatus(ctx context.Context, obj provider.AsyncResponse) (provider.JobStatus, error) {
	if obj.DocumentID == "" {
		return provider.JobStatus{}, serr.New(serr.ErrInvalidRequest, "CheckDocumentStatus", string(provider.DeepL), fmt.Errorf("Document ID not set"))
	}
	if obj.DocumentKey == "" {
		return provider.JobStatus{}, serr.New(serr.ErrInvalidRequest, "CheckDocumentStatus", string(provider.DeepL), fmt.Errorf("Document Key not set"))
	}

	data := url.Values{}
	data.Set("document_key", obj.DocumentKey)

	url := c.BaseURL.JoinPath("document", obj.DocumentID)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return provider.JobStatus{}, serr.New(serr.ErrInvalidRequest, "CheckDocumentStatus", string(provider.DeepL), fmt.Errorf("Error creating http request: %w", err))
	}
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return provider.JobStatus{}, serr.New(serr.ErrNetwork, "CheckDocumentStatus", string(provider.DeepL), err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		return provider.JobStatus{}, serr.New(serr.ErrHTTP, "CheckDocumentStatus", string(provider.DeepL), fmt.Errorf("response code %v , Trace ID: %v", res.StatusCode, res.Header.Get("X-Trace-ID")))
	}

	status := new(Status)
	jobstatus := new(provider.JobStatus)
	err = json.NewDecoder(res.Body).Decode(status)
	if err != nil {
		return provider.JobStatus{}, serr.New(serr.ErrInvalidResponse, "CheckDocumentStatus", string(provider.DeepL), fmt.Errorf("Error json decoding: %w", err))
	}

	jobstatus = &provider.JobStatus{
		Done:             false,
		Failed:           false,
		SecondsRemaining: 0,
		Message:          "",
	}

	if status.Status == "error" {
		jobstatus.Failed = true
		jobstatus.Message = status.ErrMessage
	} else if status.Status == "translating" {
		jobstatus.SecondsRemaining = status.SecondsRemaining
	} else if status.Status == "done" {
		jobstatus.Done = true
	} else if status.Status == "queued" {
		jobstatus.Message = "Queued"
	}

	if jobstatus.Failed {

		return *jobstatus, serr.New(serr.ErrProviderAPI, "CheckDocumentStatus", string(provider.DeepL), fmt.Errorf("%s", status.ErrMessage))
	}

	return *jobstatus, nil

} // expected for obj to contain docid and dockey

// TODO: manage 404 and 503 responses better
func (c *DeepLClient) GetResult(ctx context.Context, obj provider.AsyncResponse) (provider.Response, error) {
	if obj.DocumentID == "" {
		return provider.Response{}, serr.New(serr.ErrInvalidRequest, "GetResult", string(provider.DeepL), fmt.Errorf("Document ID not set"))
	}
	if obj.DocumentKey == "" {
		return provider.Response{}, serr.New(serr.ErrInvalidRequest, "GetResult", string(provider.DeepL), fmt.Errorf("Document Key not set"))
	}

	data := url.Values{}
	data.Set("document_key", obj.DocumentKey)

	url := c.BaseURL.JoinPath("document", obj.DocumentID, "result")

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return provider.Response{}, serr.New(serr.ErrInvalidRequest, "GetResult", string(provider.DeepL), fmt.Errorf("Error creating http request: %w", err))
	}
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return provider.Response{}, serr.New(serr.ErrNetwork, "GetResult", string(provider.DeepL), err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		if res.StatusCode == 404 {
			return provider.Response{}, serr.New(serr.ErrProviderAPI, "GetResult", string(provider.DeepL), errors.New("Document Not Found"))
		}
		if res.StatusCode == 503 {
			return provider.Response{}, serr.New(serr.ErrProviderAPI, "GetResult", string(provider.DeepL), errors.New("Document Already downloaded"))
		}
		return provider.Response{}, serr.New(serr.ErrHTTP, "GetResult", string(provider.DeepL), fmt.Errorf("response code %v , Trace ID: %v", res.StatusCode, res.Header.Get("X-Trace-ID")))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return provider.Response{}, serr.New(serr.ErrIO, "GetResult", string(provider.DeepL), fmt.Errorf("Error reading Document: %w", err))
	}

	return provider.Response{Binary: body}, nil

} // expected for obj to contain docid and dockey
