package config

import (
	"osrs-progress-lambda-go/wiseoldman"
	"reflect"
	"testing"
	"time"
)

func TestParseSortBy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    SortBy
		wantErr bool
	}{
		{
			name:  "Sort by Experience",
			input: "exp",
			want:  SortByExp,
		},
		{
			name:  "Sort by Efficient Hours Played",
			input: "ehp",
			want:  SortByEhp,
		},
		{
			name:  "Sort by Efficient Hours Bossing",
			input: "ehb",
			want:  SortByEhb,
		},
		{
			name:    "Invalid input",
			input:   "ehq",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseSortBy(test.input)

			if test.wantErr {
				if err == nil {
					t.Error("expected an error")
				}
				return
			}

			if err != nil && !test.wantErr {
				t.Errorf("unexpected error: %s with test: %s", err, test.name)
			}

			if got != test.want {
				t.Errorf("expected: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestParseUsernames(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "Single name",
			input: "MyCoolRsn,",
			want:  []string{"MyCoolRsn"},
		},
		{
			name:  "Multiple names",
			input: "MyCoolRsn,MyUncoolRsn,MySickRsn",
			want:  []string{"MyCoolRsn", "MyUncoolRsn", "MySickRsn"},
		},
		{
			name:    "Name is too long",
			input:   "MyNameIsTooLongForTheRsnLimit second name",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "No names entered",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Usernames need to be seperated by comma",
			input:   "username",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid input, just a comma",
			input:   ",",
			want:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseUsernames(test.input)

			if test.wantErr {
				if err == nil {
					t.Error("expected an error")
				}
			}

			if len(got) != len(test.want) {
				t.Errorf("expected: %v, got: %v", test.want, got)
			}

			for index, name := range got {
				if name != test.want[index] {
					t.Errorf("expected: %v, got: %v", test.want, got)
				}
			}
		})
	}
}

func TestParseGainsQueryMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    GainsQueryMode
		wantErr bool
	}{
		{
			name:  "period mode",
			input: "period",
			want:  GainsQueryModePeriod,
		},
		{
			name:  "range mode",
			input: "range",
			want:  GainsQueryModeRange,
		},
		{
			name:    "invalid mode",
			input:   "banana",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseGainsQueryMode(test.input)

			if test.wantErr {
				if err == nil {
					t.Error("expected an error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if got != test.want {
				t.Errorf("expected %s, got %s", test.want, got)
			}
		})
	}
}

func TestParseGainsPeriod(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    wiseoldman.GainsPeriod
		wantErr bool
	}{
		{
			name:  "Period 5 min",
			input: []string{"5min", "5m", "five_min"},
			want:  wiseoldman.GainsPeriodFiveMin,
		},
		{
			name:  "Period 1 day",
			input: []string{"day", "1day"},
			want:  wiseoldman.GainsPeriodDay,
		},
		{
			name:  "Period 1 week",
			input: []string{"week", "1week", "1 week", "1_week"},
			want:  wiseoldman.GainsPeriodWeek,
		},
		{
			name:    "Period 1 month",
			input:   []string{"month", "1month", "1 month", "1_month"},
			want:    wiseoldman.GainsPeriodMonth,
			wantErr: false,
		},
		{
			name:  "Period 1 year",
			input: []string{"year", "1year", "1_year", "1 year"},
			want:  wiseoldman.GainsPeriodYear,
		},
		{
			name:    "Invalid input",
			input:   []string{"365d", "2 years", ""},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, period := range test.input {
				got, err := parseGainsPeriod(period)

				if test.wantErr {
					if err == nil {
						t.Error("expected an error")
					}
					continue
				}

				if err != nil {
					t.Errorf("unexpected error: %s with test: %s", err, test.name)
					continue
				}

				if got != test.want {
					t.Errorf("expected: %v, got: %v", test.want, got)
				}
			}
		})
	}
}

func TestParseTimezone(t *testing.T) {
	london, loadErr := time.LoadLocation("Europe/London")
	if loadErr != nil {
		t.Fatal(loadErr)
	}

	tests := []struct {
		name    string
		input   string
		want    *time.Location
		wantErr bool
	}{
		{
			name:    "valid timezone",
			input:   "Europe/London",
			want:    london,
			wantErr: false,
		},
		{
			name:    "time zone is not defined",
			input:   "",
			want:    time.UTC,
			wantErr: false,
		},
		{
			name:    "invalid timezone",
			input:   "Gielinor/Lumbridge",
			want:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseTimezone(test.input)

			if test.wantErr {
				if err == nil {
					t.Error("expected an error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %s with test: %s", err, test.name)
			}

			if got.String() != test.want.String() {
				t.Errorf("expected: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestRequiredEnv(t *testing.T) {
	t.Setenv("MY_COOL_ENV", "SECRET")

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid env variable",
			input: "MY_COOL_ENV",
			want:  "SECRET",
		},
		{
			name:    "empty env variable",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid env variable",
			input:   "RANDOM_VARIABLE",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := requiredEnv(test.input)

			if test.wantErr {
				if err == nil {
					t.Error("expected an error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %s with test: %s", err, test.name)
			}

			if got != test.want {
				t.Errorf("expected: %v got: %v", test.want, got)
			}
		})
	}
}

func TestSuccessfulLoadFromEnv(t *testing.T) {
	env := validEnv()

	for key, value := range env {
		t.Setenv(key, value)
	}

	cfg, err := LoadFromEnv()

	if err != nil {
		t.Fatal("unexpected error")
	}

	if len(cfg.Usernames) != 3 {
		t.Errorf("expected 3 usernames, got %d", len(cfg.Usernames))
	}

	if cfg.SortBy != SortByEhp {
		t.Errorf("expected SortByEhp, got %s", cfg.SortBy)
	}

	if cfg.WebhookUrl != "https://example.com/" {
		t.Errorf("expected WebhookUrl to equal \"https://example.com/, got %s", cfg.WebhookUrl)
	}

	if cfg.GainsQueryMode != GainsQueryModeRange {
		t.Errorf("expected GainsQueryModeRange, got %s", cfg.GainsQueryMode)
	}

	if cfg.GainsPeriod != wiseoldman.GainsPeriodFiveMin {
		t.Errorf("expected GainsPeriodFiveMin, got %s", cfg.GainsPeriod)
	}

	if cfg.Timezone.String() != "Europe/London" {
		t.Errorf("expected Timezone to be utc, got %s", cfg.Timezone)
	}
}

func TestUnsuccessfulLoadFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envOverride map[string]string
	}{
		{
			name: "invalid sort by",
			envOverride: map[string]string{
				"SORT_BY": "",
			},
		},
		{
			name: "invalid sort by",
			envOverride: map[string]string{
				"SORT_BY": "WHATS A SORT BY?",
			},
		},
		{
			name: "empty username",
			envOverride: map[string]string{
				"USERNAMES": "",
			},
		},
		{
			name: "name is too long",
			envOverride: map[string]string{
				"USERNAMES": "reallyreallyreallylongname",
			},
		},
		{
			name: "invalid webhook url",
			envOverride: map[string]string{
				"WEBHOOK_URL": "",
			},
		},
		{
			name: "invalid gains query mode",
			envOverride: map[string]string{
				"GAINS_QUERY_MODE": "",
			},
		},
		{
			name: "invalid gains query mode",
			envOverride: map[string]string{
				"GAINS_QUERY_MODE": "banana",
			},
		},
		{
			name: "invalid gains period",
			envOverride: map[string]string{
				"GAINS_PERIOD": "",
			},
		},
		{
			name: "invalid gains period",
			envOverride: map[string]string{
				"GAINS_PERIOD": "42398423849",
			},
		},
		{
			name: "invalid timezone",
			envOverride: map[string]string{
				"TIMEZONE": "not a timezone",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := validEnv()

			for key, value := range test.envOverride {
				env[key] = value
			}

			for key, value := range env {
				t.Setenv(key, value)
			}

			_, err := LoadFromEnv()
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func validEnv() map[string]string {
	return map[string]string{
		"USERNAMES":        "Testuser,testeruser2,testuser3",
		"SORT_BY":          "ehp",
		"WEBHOOK_URL":      "https://example.com/",
		"GAINS_QUERY_MODE": "range",
		"GAINS_PERIOD":     "5min",
		"TIMEZONE":         "Europe/London",
		"EMBED_COLOUR":     "",
		"THUMBNAIL_URL":    "",
		"IMAGE_URL":        "",
	}
}

func TestRemoveDuplicateUsernames(t *testing.T) {
	tests := []struct {
		name      string
		usernames []string
		want      []string
	}{
		{
			name: "Remove One duplicate username",
			usernames: []string{
				"user1", "user2", "user3", "user1",
			},
			want: []string{
				"user1", "user2", "user3",
			},
		},
		{
			name: "Remove Multiple duplicate usernames",
			usernames: []string{
				"user1", "user2", "user3", "user1", "user1", "user2", "user3", "user1",
			},
			want: []string{
				"user1", "user2", "user3",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := removeDuplicateUsernames(test.usernames)

			reflect.DeepEqual(r, test.want)
		})
	}
}
