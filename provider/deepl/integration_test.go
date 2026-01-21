//go:build integration

package deepl

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	format "github.com/o0n1x/mass-translate-package/format"
	lang "github.com/o0n1x/mass-translate-package/lang"
	provider "github.com/o0n1x/mass-translate-package/provider"
)

func TestDeeplIntegrationText(t *testing.T) {
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

func TestDeeplIntegrationFile(t *testing.T) {

	input_file_name := "test_files/Houseki_no_Kuni_s1e1.srt"
	output_file_name := "test_files/Houseki_no_Kuni_s1e1_ar.srt"

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
	resp, err := client.Translate(context.Background(), provider.Request{
		ReqType: format.File,
		Text:    []string{InputFile.Name()},
		From:    lang.English,
		To:      lang.Arabic,
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

		t.Logf("Status of Request: %v", status.Status)

		if status.Status == "done" {
			break
		} else if status.Status == "error" {
			t.Fatalf("Status Error: %v", status.ErrMessage)

		} else if status.Status == "translating" {
			t.Logf("Time remaining till completion: %v", status.SecondsRemaining)
		}

	}

	file, err := client.GetResult(context.Background(), resp)
	if err != nil {
		t.Fatalf("Error Getting result: %v", err)
	}

	err = os.WriteFile(output_file_name, file, 0644)
	if err != nil {
		t.Fatalf("Error writing result: %v", err)
	}

}
