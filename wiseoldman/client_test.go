package wiseoldman

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		inputUrl string
		want     *Client
		wantErr  bool
	}{
		{
			name:     "valid NewClient",
			inputUrl: ApiUrl,
			want: &Client{
				httpClient: &http.Client{},
				baseUrl: &url.URL{
					Scheme: "https",
					Host:   "api.wiseoldman.net",
					Path:   "/v2/players/",
				},
			},
		},
		{
			name:     "invalid NewClient Empty Url",
			inputUrl: "",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewClient(test.inputUrl)

			if test.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got.baseUrl.String() != test.want.baseUrl.String() {
				t.Errorf("expected baseUrl: %s, got: %s", test.want.baseUrl.String(), got.baseUrl.String())
			}
		})
	}
}

func TestSendUpdateForPlayerSuccess(t *testing.T) {
	body, err := os.ReadFile("testdata/player_update_success.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected post got %s", r.Method)
		}

		if r.URL.Path != "/username" {
			t.Errorf("expected /username got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(body)
	}))

	defer server.Close()

	client, err := NewClient(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.SendUpdateForPlayer("username")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got.LastChangedAt == nil {
		t.Errorf("expected LastChangedAt got nil")
	}

	if got.LastChangedAt.Format(DateLayout) != "2026-02-25T13:00:53.033Z" {
		t.Errorf("expected LastChangedAt got %s", got.LastChangedAt)
	}
}

func TestSendUpdateForPlayerError(t *testing.T) {
	var apiError = ApiError{
		Code:    "HIGHSCORES_USERNAME_NOT_FOUND",
		Message: "Player not found on the hiscores",
	}

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("expected post got %s", request.Method)
		}

		if request.URL.Path != "/username" {
			t.Errorf("expected /username got %s", request.URL.Path)
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(responseWriter).Encode(apiError)
		if err != nil {
			t.Errorf("unexpected write error: %v", err)
		}
	}))

	defer server.Close()

	client, err := NewClient(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.SendUpdateForPlayer("username")
	if err == nil {
		t.Error("expected error")
	}

	var gotApiError ApiError
	if !errors.As(err, &gotApiError) {
		t.Errorf("expected ApiError got %T: %v", err, err)
	}
	if gotApiError.Code != apiError.Code {
		t.Errorf("expected ApiError.Code %q, got %q", apiError.Code, gotApiError.Code)
	}

	if gotApiError.Message != apiError.Message {
		t.Errorf("expected ApiError.Message %q, got %q", apiError.Message, gotApiError.Message)
	}

	if got.LastChangedAt != nil {
		t.Errorf("expected LastChangedAt got nil")
	}
}

func TestPlayerGainsSuccess(t *testing.T) {
	const wantedOverallExpGain = 85532
	const wantedEhbStart = 12.84

	body, err := os.ReadFile("testdata/player_gains_success.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("expected Get got %s", request.Method)
		}

		if request.URL.Path != "/username/gained" {
			t.Errorf("expected /username/gained got %s", request.URL.Path)
		}

		query := request.URL.Query()
		if got := query.Get("period"); got == "" {
			t.Errorf("expected period got empty string")
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusOK)

		_, _ = responseWriter.Write(body)
	}))

	defer server.Close()

	client, err := NewClient(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	queryParams := PlayerGainsParams{
		GainsPeriod: new(GainsPeriodDay),
		StartDate:   nil,
		EndDate:     nil,
	}

	username := "username"
	got, err := client.GetPlayerData(username, &queryParams)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got.PlayerName != username {
		t.Errorf("expected %s got %s", username, got.PlayerName)
	}

	if got.StartsAt == nil {
		t.Errorf("expected startsAt to be set got nil")
	}

	if got.EndsAt == nil {
		t.Errorf("expected endsAt to be set got nil")
	}

	if got.PlayerData.Skills == nil {
		t.Errorf("expected skills to be set got nil")
	}

	if got.PlayerData.Bosses == nil {
		t.Errorf("expected bosses to be set got nil")
	}

	if got.PlayerData.Activities == nil {
		t.Errorf("expected activites to be set got nil")
	}

	if got.PlayerData.Computed == nil {
		t.Errorf("expected computed to be set got nil")
	}

	if got.PlayerData.Skills["overall"].Experience.Gained != wantedOverallExpGain {
		t.Errorf("expected %d got %d", wantedOverallExpGain, got.PlayerData.Skills["overall"].Experience.Gained)
	}

	if got.PlayerData.Bosses["abyssal_sire"].Ehb.StartedAt != wantedEhbStart {
		t.Errorf("expected %f got %f", wantedEhbStart, got.PlayerData.Bosses["abyssal_sire"].Ehb.StartedAt)
	}
}

func TestPlayerGainsError(t *testing.T) {
	var apiError = ApiError{
		Code:    "UNEXPECTED_ERROR",
		Message: "Player not found.",
	}

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("expected Get got %s", request.Method)
		}

		if request.URL.Path != "/username/gained" {
			t.Errorf("expected /username/gained got %s", request.URL.Path)
		}

		query := request.URL.Query()
		if got := query.Get("period"); got == "" {
			t.Errorf("expected period got empty string")
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(responseWriter).Encode(apiError)
		if err != nil {
			t.Errorf("unexpected write error: %v", err)
		}
	}))

	defer server.Close()

	client, err := NewClient(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	queryParams := PlayerGainsParams{
		GainsPeriod: new(GainsPeriodDay),
		StartDate:   nil,
		EndDate:     nil,
	}

	username := "username"
	_, err = client.GetPlayerData(username, &queryParams)
	if err == nil {
		t.Error("expected error, got nil")
	}

	var gotApiError ApiError
	if !errors.As(err, &gotApiError) {
		t.Errorf("expected ApiError got %T: %v", err, err)
	}
	if gotApiError.Code != apiError.Code {
		t.Errorf("expected ApiError.Code %q, got %q", apiError.Code, gotApiError.Code)
	}

	if gotApiError.Message != apiError.Message {
		t.Errorf("expected ApiError.Message %q, got %q", apiError.Message, gotApiError.Message)
	}
}
