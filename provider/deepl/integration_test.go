//go:build integration

package deepl

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	format "github.com/o0n1x/sublate-go/format"
	lang "github.com/o0n1x/sublate-go/lang"
	provider "github.com/o0n1x/sublate-go/provider"
)

func TestDeeplIntegrationText(t *testing.T) {
	godotenv.Load("../../.env")
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Fatal("DEEPL_API_KEY not set")
	}

	deeplclient, err := provider.GetClient(provider.DeepL, apiKey)
	if err != nil {
		t.Fatal(err)
	}
	client := deeplclient.(provider.SyncClient)
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

func TestDeeplIntegrationFile(t *testing.T) {

	input_file_name := "test_files/inputTest.srt"
	output_file_name := "test_files/outputTest.srt"
	filename := "inputTest.srt"

	godotenv.Load("../../.env")
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Fatal("DEEPL_API_KEY not set")
	}

	generalizedclient, err := provider.GetClient(provider.DeepL, apiKey)
	if err != nil {
		t.Fatal(err)
	}

	client := generalizedclient.(*DeepLClient)

	//verify file exists

	InputFile, err := os.Open(input_file_name)
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(InputFile)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.AsyncTranslate(context.Background(), provider.Request{
		ReqType:  format.File,
		Binary:   data,
		FileName: filename,
		From:     lang.English,
		To:       lang.Arabic,
	})
	if err != nil {
		t.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
		status, err := client.CheckStatus(context.Background(), resp)

		if err != nil {
			t.Fatalf("Error checking status: %v", err)
		}

		if status.Done {
			break
		} else if status.Failed {
			t.Fatalf("Status Error: %v", status.Message)

		} else {
			t.Logf("Time remaining till completion: %v", status.SecondsRemaining)
		}

	}

	file, err := client.GetResult(context.Background(), resp)
	if err != nil {
		t.Fatalf("Error Getting result: %v", err)
	}

	err = os.WriteFile(output_file_name, file.Binary, 0644)
	if err != nil {
		t.Fatalf("Error writing result: %v", err)
	}

}
