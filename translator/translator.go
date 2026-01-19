package translator

import (
	"context"

	provider "github.com/o0n1x/mass-translate-package/provider"
)

func Translate(ctx context.Context, req provider.Request, client provider.Client) (provider.Response, error) {
	panic("not implemented")
}

// wrapper function that parralelizes translation based on the provider
// by design this will wait for all batch to be completed and return all results/ errors even if its async
// TODO: make it possible to jst return async results and get status of the async results
func BatchTranslate(ctx context.Context, req []provider.Request, client provider.Client) (provider.Response, error) {
	panic("not implemented")
}
