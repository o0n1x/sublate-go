package translator

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	format "github.com/o0n1x/Sublate-go/format"
	lang "github.com/o0n1x/Sublate-go/lang"
	provider "github.com/o0n1x/Sublate-go/provider"
)

func TestTranslatorIntegration(t *testing.T) {
	godotenv.Load("../.env")
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Fatal("DEEPL_API_KEY not set")
	}

	client, err := provider.GetClient(provider.DeepL, apiKey)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := Translate(context.Background(), provider.Request{
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
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Text) == 0 {
		t.Fatal("empty response")
	}
	// real API returns real translation
	t.Logf("Got: %s", resp.Text)
}

func TestBatchTranslatorIntegration(t *testing.T) {
	godotenv.Load("../.env")
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Fatal("DEEPL_API_KEY not set")
	}

	input_file_name := "../provider/deepl/test_files/inputTest.srt"
	output_file_name := "../provider/deepl/test_files/outputTest.srt"

	InputFile, err := os.Open(input_file_name)
	if err != nil {
		t.Fatal(err)
	}

	client, err := provider.GetClient(provider.DeepL, apiKey)
	if err != nil {
		t.Fatal(err)
	}
	var batcherr []error
	resp, batcherr := BatchTranslate(context.Background(), []provider.Request{
		provider.Request{
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
		}, provider.Request{
			ReqType: format.File,
			Text:    []string{InputFile.Name()},
			To:      lang.Arabic,
		},
	}, client)
	if batcherr[0] != nil {
		t.Logf("Error for number 0: %v", err)
	} else {
		t.Logf("Got for number 0: %v", resp[0].Text)
	}
	if batcherr[1] != nil {
		t.Logf("Error for number 1: %v", batcherr[1])
	} else {
		err = os.WriteFile(output_file_name, resp[1].Binary, 0644)
		if err != nil {
			t.Fatalf("Error writing result: %v", err)
		}
	}
}
