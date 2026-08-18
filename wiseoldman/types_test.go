package wiseoldman

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestPlayerGainsParams_Validate(t *testing.T) {
	tests := []struct {
		name      string
		input     PlayerGainsParams
		wantError bool
	}{
		{
			name: "GainsPeriod",
			input: PlayerGainsParams{
				GainsPeriod: new(GainsPeriodYear),
				StartDate:   nil,
				EndDate:     nil,
			},
			wantError: false,
		},
		{
			name: "StartDate & EndDate",
			input: PlayerGainsParams{
				GainsPeriod: nil,
				StartDate:   new(time.Date(2026, 8, 1, 15, 30, 3, 0, time.UTC)),
				EndDate:     new(time.Date(2026, 8, 2, 15, 30, 3, 0, time.UTC)),
			},
			wantError: false,
		},
		{
			name: "Invalid Params",
			input: PlayerGainsParams{
				GainsPeriod: nil,
				StartDate:   nil,
				EndDate:     nil,
			},
			wantError: true,
		},
		{
			name: "One date set",
			input: PlayerGainsParams{
				GainsPeriod: nil,
				StartDate:   new(time.Date(2026, 8, 1, 15, 30, 3, 0, time.UTC)),
				EndDate:     nil,
			},
			wantError: true,
		},
		{
			name: "end date is before the start date",
			input: PlayerGainsParams{
				GainsPeriod: nil,
				StartDate:   new(time.Date(2026, 8, 2, 15, 30, 3, 0, time.UTC)),
				EndDate:     new(time.Date(2026, 8, 1, 15, 30, 3, 0, time.UTC)),
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.input.Validate()

			if test.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
		})
	}
}

func TestPlayerGainsParams_QueryValues(t *testing.T) {
	tests := []struct {
		name  string
		input PlayerGainsParams
		want  url.Values
	}{
		{
			name: "set gains period",
			input: PlayerGainsParams{
				GainsPeriod: new(GainsPeriodYear),
			},
			want: url.Values{
				"period": []string{"year"},
			},
		},
		{
			name: "set start date and end date",
			input: PlayerGainsParams{
				StartDate: new(time.Date(2026, 8, 1, 15, 30, 3, 0, time.UTC)),
				EndDate:   new(time.Date(2026, 8, 2, 15, 30, 3, 0, time.UTC)),
			},
			want: url.Values{
				"startDate": []string{"2026-08-01T15:30:03.000Z"},
				"endDate":   []string{"2026-08-02T15:30:03.000Z"},
			},
		},
		{
			name:  "empty params",
			input: PlayerGainsParams{},
			want:  url.Values{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.QueryValues()

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("expected %v, got %v", test.want, got)
			}
		})
	}
}
