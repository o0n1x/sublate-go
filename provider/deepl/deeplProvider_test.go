package deepl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	lang "github.com/o0n1x/sublate-go/lang"
)

func TestTranslateText(t *testing.T) {
	cases := map[string]struct {
		text  string
		from  lang.Language
		to    lang.Language
		trans string
	}{
		"simple":     {"hello", lang.English, lang.German, "hallo"},
		"empty_text": {"", lang.English, lang.German, ""},
		"arabic":     {"hello", lang.English, lang.Arabic, "مرحبًا"},
		"korean":     {"hello", lang.English, lang.Korean, "안녕하세요"},
		"Chinese":    {"hello", lang.English, lang.ChineseSimplified, "你好"},
		"AutoDetect": {"hello", lang.AutoDetect, lang.German, "hallo"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// dummy api server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/translate" {
					t.Errorf("wrong path: %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"translations":[{"text":"%s"}]}`, tc.trans)
			}))
			defer server.Close()

			u, _ := url.Parse(server.URL)
			client := &DeepLClient{
				Client:  server.Client(),
				BaseURL: u.JoinPath(APIVersion),
				APIKey:  "test-key",
			}

			resp, err := client.translateText(context.Background(), []string{tc.text}, tc.from, tc.to)

			if err != nil {
				t.Fatal(err)
			}
			if resp.Text[0] != tc.trans {
				t.Errorf("got %s, want %s", resp.Text, tc.trans)
			}

		})
	}
}
