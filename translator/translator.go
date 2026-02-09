package translator

import (
	"context"
	"fmt"
	"time"

	provider "github.com/o0n1x/mass-translate-package/provider"
	deepl "github.com/o0n1x/mass-translate-package/provider/deepl"
)

// Translate is the main entry point for all translations.
// Handles async providers (DeepL files) internally - always returns sync response.
// Use this, not client.Translate() directly.
func Translate(ctx context.Context, req provider.Request, client provider.Client) (provider.Response, error) {
	switch client.Name() {
	case provider.DeepL:
		return translateDeepl(ctx, req, client)
	default:
		return provider.Response{}, fmt.Errorf("Invalid Client type")
	}
}

func translateDeepl(ctx context.Context, req provider.Request, client provider.Client) (provider.Response, error) {
	deeplC := client.(*deepl.DeepLClient)
	res, err := deeplC.Translate(ctx, req)
	if err != nil {
		return provider.Response{}, fmt.Errorf("Error translating using DeepL provider: %v", err)
	}
	switch res.ResType { //this approach is abit too coupled and not generalized ,though, its fine for now TODO: find a better approach
	case provider.Sync:
		//sync jst return
		return res, nil

	case provider.ASync:
		//async wait for completion
		status, err := deeplC.CheckStatus(ctx, res)
		if err != nil {
			return provider.Response{}, fmt.Errorf("Error translating using DeepL provider: %v", err)
		}
		for status.Status != "done" {
			if status.Status == "error" {
				return provider.Response{}, fmt.Errorf("Translation failed: %v", status.ErrMessage)
			}

			//keeps an eye for ctx cancellation, if it closes mid translation it returns
			select {
			case <-ctx.Done():
				return provider.Response{}, ctx.Err()
			case <-time.After(time.Second):
			}

			status, err = deeplC.CheckStatus(ctx, res)
			if err != nil {
				return provider.Response{}, fmt.Errorf("Error translating using DeepL provider: %v", err)
			}
		}
		translation, err := deeplC.GetResult(ctx, res)
		if err != nil {
			return provider.Response{}, fmt.Errorf("Error translating using DeepL provider: %v", err)
		}

		return provider.Response{
			ResType: provider.Sync,
			Binary:  translation,
		}, nil

	default:
		return provider.Response{}, fmt.Errorf("Error translating using DeepL provider: Invalid request type")
	}

}

// wrapper function that parralelizes translation based on the provider
// by design this will wait for all batch to be completed and return all results/ errors even if its async
// TODO: make it possible to jst return async results and get status of the async results
func BatchTranslate(ctx context.Context, req []provider.Request, client provider.Client) ([]provider.Response, []error) {
	var responses = make([]provider.Response, len(req))
	var errs = make([]error, len(req))

	for i, request := range req {
		res, err := Translate(ctx, request, client)
		if err != nil {
			errs[i] = fmt.Errorf("Error Occured at request number %v: %v", i, err)
			continue
		}
		responses[i] = res
	}
	return responses, errs

}

// comments for usage: a single pdf uses ALOT of chars (10-50k+) if you want to translate pdfs i suggest extracting text and translate that text.
