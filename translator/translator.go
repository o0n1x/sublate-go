package translator

import (
	"context"
	"fmt"
	"time"

	serr "github.com/o0n1x/sublate-go/errors"
	sformat "github.com/o0n1x/sublate-go/format"
	provider "github.com/o0n1x/sublate-go/provider"
)

//TODO: we need to find a way to

// Translate is the main entry point for all translations.
// Handles async and sync providers  internally - always returns sync response.
// Use this, not client.Translate() directly.
func Translate(ctx context.Context, req provider.Request, client provider.Client) (provider.Response, error) {
	switch req.ReqType {
	case sformat.File:
		asyncC, ok := client.(provider.AsyncClient)
		if !ok {
			return provider.Response{}, serr.New(serr.ErrInvalidRequest, "Translate", "", fmt.Errorf("client does not support file translation"))
		}
		return translateAsyncComplete(ctx, req, asyncC)
	case sformat.Text:
		syncC, ok := client.(provider.SyncClient)
		if !ok {
			return provider.Response{}, serr.New(serr.ErrInvalidRequest, "Translate", "", fmt.Errorf("client does not support text translation"))
		}
		return translateSync(ctx, req, syncC)
	default:
		return provider.Response{}, serr.New(serr.ErrInvalidRequest, "Translate", "", fmt.Errorf("invalid request type"))

	}
}

func translateSync(ctx context.Context, req provider.Request, client provider.SyncClient) (provider.Response, error) {
	res, err := client.Translate(ctx, req)
	if err != nil {
		return provider.Response{}, err // all errors returned from deepl is wrapped as serr
	}
	return res, nil
}

func translateAsync(ctx context.Context, req provider.Request, client provider.AsyncClient) (provider.AsyncResponse, error) {
	res, err := client.AsyncTranslate(ctx, req)
	if err != nil {
		return provider.AsyncResponse{}, err // all errors returned from deepl is wrapped as serr
	}
	return res, nil
}

func translateAsyncComplete(ctx context.Context, req provider.Request, client provider.AsyncClient) (provider.Response, error) {
	res, err := client.AsyncTranslate(ctx, req)
	if err != nil {
		return provider.Response{}, err // all errors returned from deepl is wrapped as serr
	}
	status, err := client.CheckStatus(ctx, res)
	if err != nil {
		return provider.Response{}, err
	}

	for !status.Done {

		//keeps an eye for ctx cancellation, if it closes mid translation it returns
		select {
		case <-ctx.Done():
			return provider.Response{}, serr.New(serr.ErrNetwork, "Translate", string(client.Name()), ctx.Err())
		case <-time.After(time.Second):
		}

		status, err = client.CheckStatus(ctx, res)
		if err != nil {
			return provider.Response{}, err
		}
	}
	translation, err := client.GetResult(ctx, res)
	if err != nil {
		return provider.Response{}, err
	}

	return translation, nil
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
			errs[i] = err
			continue
		}
		responses[i] = res
	}
	return responses, errs

}

// comments for usage: a single pdf uses ALOT of chars (10-50k+) if you want to translate pdfs i suggest extracting text and translate that text.
