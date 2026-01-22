package deepl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: baseURL,
		APIKey:  apiKey,
		IsFree:  isFreeAccount(apiKey),
	}
}

// will verify the input like from/to lang is valid and use the appropriate helper function to get translation
func (c *DeepLClient) Translate(ctx context.Context, req provider.Request) (provider.Response, error) {
	//validate lang
	if req.From == "" {
		req.From = lang.AutoDetect
	}
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
	file, err := os.Open(docPath)
	if err != nil {
		return provider.Response{}, fmt.Errorf("Invalid filepath: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return provider.Response{}, fmt.Errorf("Error writing file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return provider.Response{}, fmt.Errorf("Error Copying File: %v", err)
	}

	if from != lang.AutoDetect {
		err = writer.WriteField("source_lang", from.String())
		if err != nil {
			return provider.Response{}, fmt.Errorf("Error writing source_lang to request body")
		}
	}

	err = writer.WriteField("target_lang", to.String())
	if err != nil {
		return provider.Response{}, fmt.Errorf("Error writing target_lang to request body")
	}

	writer.Close()

	url := c.BaseURL.JoinPath("/document")
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), body)
	if err != nil {
		return provider.Response{}, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := c.Client.Do(req)
	if err != nil {
		return provider.Response{}, fmt.Errorf("Error sending request to server API: %v", err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		return provider.Response{}, fmt.Errorf("Document upload Error: HTTP Error %v \n", res.StatusCode)
	}

	document := new(Documents)

	err = json.NewDecoder(res.Body).Decode(document)
	if err != nil {
		return provider.Response{}, err
	}

	return provider.Response{
		ResType:     provider.ASync,
		DocumentID:  document.DocID,
		DocumentKey: document.DocKey,
	}, nil

}

func (c *DeepLClient) CheckStatus(ctx context.Context, obj provider.Response) (Status, error) {
	if obj.DocumentID == "" {
		return Status{}, fmt.Errorf("Document ID not set")
	}
	if obj.DocumentKey == "" {
		return Status{}, fmt.Errorf("Document Key not set")
	}

	data := url.Values{}
	data.Set("document_key", obj.DocumentKey)

	url := c.BaseURL.JoinPath("document", obj.DocumentID)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return Status{}, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return Status{}, fmt.Errorf("Error sending request to server API: %v", err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		return Status{}, fmt.Errorf("Check status Error: HTTP Error %v \n", res.StatusCode)
	}

	status := new(Status)

	err = json.NewDecoder(res.Body).Decode(status)
	if err != nil {
		return Status{}, err
	}

	if status.Status == "error" {
		return *status, fmt.Errorf("Error from Deepl API: %v", status.ErrMessage)
	}

	return *status, nil

} // expected for obj to contain docid and dockey

func (c *DeepLClient) GetResult(ctx context.Context, obj provider.Response) ([]byte, error) {
	if obj.DocumentID == "" {
		return nil, fmt.Errorf("Document ID not set")
	}
	if obj.DocumentKey == "" {
		return nil, fmt.Errorf("Document Key not set")
	}

	data := url.Values{}
	data.Set("document_key", obj.DocumentKey)

	url := c.BaseURL.JoinPath("document", obj.DocumentID, "result")

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request to server API: %v", err)
	}
	defer res.Body.Close()

	ok := http.StatusOK <= res.StatusCode && res.StatusCode < http.StatusMultipleChoices
	if !ok {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("Deepl API Error: Document Not Found")
		}
		if res.StatusCode == 503 {
			return nil, fmt.Errorf("Deepl API Error: Document Already downloaded")
		}
		return nil, fmt.Errorf("Check status Error: HTTP Error %v \n", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading Document: %v", err)
	}

	return body, nil

} // expected for obj to contain docid and dockey
