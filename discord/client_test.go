package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Client
		wantErr bool
	}{
		{
			name:  "valid",
			input: "https://discord.com/api/webhooks/example/example",
			want: &Client{
				webhookUrl: &url.URL{
					Scheme: "https",
					Host:   "discord.com",
					Path:   "/api/webhooks/example/example",
				},
				httpClient: &http.Client{},
			},
		},
		{
			name:    "invalid",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewClient(test.input)
			if test.wantErr {
				if err == nil {
					t.Errorf("NewClient(%q) want error", test.input)
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient(%q) got error: %v", test.input, err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewClient(%q) = %+v, want %+v", test.input, got, test.want)
			}
		})
	}
}

func TestClient_SendSuccess(t *testing.T) {
	requestBody, err := os.ReadFile("testdata/embed_request_body.json")
	if err != nil {
		t.Fatal(err)
	}

	var mockEmbeds struct {
		Embeds []Embed `json:"embeds"`
	}

	if err := json.Unmarshal(requestBody, &mockEmbeds); err != nil {
		t.Fatalf("error unmarshaling embed request body %v", err)
	}

	body, err := json.Marshal(WebhookBody{
		Embeds: mockEmbeds.Embeds,
	})
	if err != nil {
		t.Errorf("error marshaling embed request body %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method post, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)

		_, _ = w.Write(body)
	}))

	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Error("error creating client")
	}

	_, err = client.Send(mockEmbeds.Embeds[0])
	if err != nil {
		t.Errorf("error sending request: %v", err)
	}
}

func TestClient_SendFail(t *testing.T) {
	requestBody, err := os.ReadFile("testdata/embed_request_body.json")
	if err != nil {
		t.Fatal(err)
	}

	var mockEmbeds struct {
		Embeds []Embed `json:"embeds"`
	}

	if err := json.Unmarshal(requestBody, &mockEmbeds); err != nil {
		t.Fatalf("error unmarshaling embed request body %v", err)
	}

	body, err := json.Marshal(WebhookBody{
		Embeds: mockEmbeds.Embeds,
	})
	if err != nil {
		t.Errorf("error marshaling embed request body %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method post, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		_, _ = w.Write(body)
	}))

	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Error("error creating client")
	}

	_, err = client.Send(mockEmbeds.Embeds[0])
	if err == nil {
		t.Errorf("expected error got: %v", err)
	}
}
