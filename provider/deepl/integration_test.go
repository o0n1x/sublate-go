//go:build integration

package deepl

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	format "github.com/o0n1x/mass-translate-package/format"
	lang "github.com/o0n1x/mass-translate-package/lang"
	provider "github.com/o0n1x/mass-translate-package/provider"
)

func TestDeeplIntegration(t *testing.T) {
	godotenv.Load("../../.env")
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Fatal("DEEPL_API_KEY not set")
	}

	client, err := provider.GetClient(provider.DeepL, apiKey)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Translate(context.Background(), provider.Request{
		ReqType: format.Text,
		Text: []string{
			"Hello, how are you today?",
			"The weather is beautiful outside.",
			"I would like to order a coffee please.",
			"Thank you for your help with this project.",
			"See you tomorrow at the meeting.",
		},
		From: lang.English,
		To:   lang.Japanese,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Text) == 0 {
		t.Fatal("empty response")
	}
	// real API returns real translation
	t.Logf("Got: %s", resp.Text)
}
