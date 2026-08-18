package main

import (
	"encoding/json"
	"errors"
	"os"
	"osrs-progress-lambda-go/config"
	"osrs-progress-lambda-go/wiseoldman"
	"reflect"
	"testing"
	"time"
)

func TestSort(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    *config.SortBy
		gains     []wiseoldman.PlayerGains
		wantOrder []string
	}{
		{
			name:   "sort by EXP descending",
			sortBy: new(config.SortByExp),
			gains: mockPlayerGains(t,
				mockGainValue{playerName: "fourth", exp: 250, ehp: 40, ehb: 4},
				mockGainValue{playerName: "second", exp: 750, ehp: 20, ehb: 2},
				mockGainValue{playerName: "first", exp: 1000, ehp: 10, ehb: 1},
				mockGainValue{playerName: "fifth", exp: 100, ehp: 50, ehb: 5},
				mockGainValue{playerName: "third", exp: 500, ehp: 30, ehb: 3},
			),
			wantOrder: []string{"first", "second", "third", "fourth", "fifth"},
		},
		{
			name:   "sort by EHP descending",
			sortBy: new(config.SortByEhp),
			gains: mockPlayerGains(t,
				mockGainValue{playerName: "third", exp: 500, ehp: 30, ehb: 3},
				mockGainValue{playerName: "fifth", exp: 1000, ehp: 10, ehb: 1},
				mockGainValue{playerName: "first", exp: 100, ehp: 50, ehb: 5},
				mockGainValue{playerName: "second", exp: 250, ehp: 40, ehb: 4},
				mockGainValue{playerName: "fourth", exp: 750, ehp: 20, ehb: 2},
			),
			wantOrder: []string{"first", "second", "third", "fourth", "fifth"},
		},
		{
			name:   "sort by EHB descending",
			sortBy: new(config.SortByEhb),
			gains: mockPlayerGains(t,
				mockGainValue{playerName: "third", exp: 500, ehp: 30, ehb: 3},
				mockGainValue{playerName: "fifth", exp: 1000, ehp: 10, ehb: 1},
				mockGainValue{playerName: "sixth", exp: 1000, ehp: 10, ehb: 0.5},
				mockGainValue{playerName: "first", exp: 100, ehp: 50, ehb: 5},
				mockGainValue{playerName: "second", exp: 250, ehp: 40, ehb: 4},
				mockGainValue{playerName: "fourth", exp: 750, ehp: 20, ehb: 2},
			),
			wantOrder: []string{"first", "second", "third", "fourth", "fifth", "sixth"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := sortPlayerGains(test.sortBy, test.gains)
			if got == nil {
				t.Fatal("expected sorted gains, got nil")
			}

			gotOrder := playerNames(got)
			if !reflect.DeepEqual(gotOrder, test.wantOrder) {
				t.Errorf("expected order %v, got %v", test.wantOrder, gotOrder)
			}
		})
	}
}

type mockGainValue struct {
	playerName string
	exp        int
	ehp        float64
	ehb        float64
}

func mockPlayerGains(t *testing.T, values ...mockGainValue) []wiseoldman.PlayerGains {
	t.Helper()

	gains := make([]wiseoldman.PlayerGains, len(values))
	for i, value := range values {
		gain := mockPlayerGain(t)
		gain.PlayerName = value.playerName

		overall := gain.PlayerData.Skills["overall"]
		overall.Experience.Gained = value.exp
		gain.PlayerData.Skills["overall"] = overall

		ehp := gain.PlayerData.Computed["ehp"]
		ehp.Value.Gained = value.ehp
		gain.PlayerData.Computed["ehp"] = ehp

		ehb := gain.PlayerData.Computed["ehb"]
		ehb.Value.Gained = value.ehb
		gain.PlayerData.Computed["ehb"] = ehb

		gains[i] = gain
	}

	return gains
}

func mockPlayerGain(t *testing.T) wiseoldman.PlayerGains {
	t.Helper()

	body, err := os.ReadFile("discord/testdata/player_gains.json")
	if err != nil {
		t.Fatalf("error reading player gains fixture: %v", err)
	}

	var gains wiseoldman.PlayerGains
	if err := json.Unmarshal(body, &gains); err != nil {
		t.Fatalf("error unmarshaling player gains fixture: %v", err)
	}

	if gains.PlayerData.Skills == nil {
		t.Fatal("fixture missing skills")
	}

	if gains.PlayerData.Computed == nil {
		t.Fatal("fixture missing computed metrics")
	}

	if _, ok := gains.PlayerData.Skills["overall"]; !ok {
		t.Fatal("fixture missing overall skill")
	}

	if _, ok := gains.PlayerData.Computed["ehp"]; !ok {
		t.Fatal("fixture missing ehp computed metric")
	}

	if _, ok := gains.PlayerData.Computed["ehb"]; !ok {
		t.Fatal("fixture missing ehb computed metric")
	}

	return gains
}

func TestSortPlayerGainsReturnsErrorWhenSortByIsNil(t *testing.T) {
	_, err := sortPlayerGains(nil, mockPlayerGains(t,
		mockGainValue{playerName: "player", exp: 1, ehp: 1, ehb: 1},
	))

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSortPlayerGainsDoesNotMutateInput(t *testing.T) {
	gains := mockPlayerGains(t,
		mockGainValue{playerName: "low", exp: 1, ehp: 1, ehb: 1},
		mockGainValue{playerName: "high", exp: 2, ehp: 2, ehb: 2},
	)

	_, err := sortPlayerGains(new(config.SortByExp), gains)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotOrder := playerNames(gains)
	wantOrder := []string{"low", "high"}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Errorf("expected original order %v, got %v", wantOrder, gotOrder)
	}
}

func TestPlayerGainsFromPlayers(t *testing.T) {
	players := []Player{
		{Username: "first", PlayerGains: mockPlayerGains(t, mockGainValue{playerName: "first", exp: 1})[0]},
		{Username: "failed", Err: errors.New("failed")},
		{Username: "no-gains", PlayerGains: mockPlayerGains(t, mockGainValue{playerName: "no-gains"})[0]},
		{Username: "second", PlayerGains: mockPlayerGains(t, mockGainValue{playerName: "second", exp: 2})[0]},
	}

	got := playerGainsFromPlayers(players)
	gotOrder := playerNames(got)
	wantOrder := []string{"first", "second"}

	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Errorf("expected gains %v, got %v", wantOrder, gotOrder)
	}
}

func TestFilter(t *testing.T) {
	period := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	before := period.Add(-time.Hour)
	after := period.Add(time.Hour)

	players := &Players{Players: []Player{
		{Username: "failed", Err: errors.New("failed"), LastChangedAt: &after},
		{Username: "missing-date"},
		{Username: "too-old", LastChangedAt: &before},
		{Username: "same-time", LastChangedAt: &period},
		{Username: "newer", LastChangedAt: &after},
	}}

	got, err := filter(players, &period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotOrder := playerUsernames(got.Players)
	wantOrder := []string{"same-time", "newer"}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Errorf("expected players %v, got %v", wantOrder, gotOrder)
	}
}

func TestFilterReturnsErrorForNilPlayers(t *testing.T) {
	_, err := filter(nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveGainsPeriodFromConfigUsesPeriodMode(t *testing.T) {
	cfg := config.Config{
		GainsQueryMode: config.GainsQueryModePeriod,
		GainsPeriod:    wiseoldman.GainsPeriodWeek,
		Timezone:       time.UTC,
	}

	got := resolveGainsPeriodFromConfig(&cfg, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	if got.GainsPeriod == nil {
		t.Fatal("expected GainsPeriod to be set")
	}

	if *got.GainsPeriod != wiseoldman.GainsPeriodWeek {
		t.Errorf("expected gains period %q, got %q", wiseoldman.GainsPeriodWeek, *got.GainsPeriod)
	}

	if got.StartDate != nil || got.EndDate != nil {
		t.Errorf("expected start/end dates to be nil")
	}
}

func TestResolveGainsPeriodFromConfigBuildsDateRange(t *testing.T) {
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	cfg := config.Config{
		GainsQueryMode: config.GainsQueryModeRange,
		GainsPeriod:    wiseoldman.GainsPeriodWeek,
		Timezone:       time.UTC,
	}

	got := resolveGainsPeriodFromConfig(&cfg, now)

	if got.StartDate == nil {
		t.Fatal("expected StartDate to be set")
	}

	if got.EndDate == nil {
		t.Fatal("expected EndDate to be set")
	}

	wantStart := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	if !got.StartDate.Equal(wantStart) {
		t.Errorf("expected start date %v, got %v", wantStart, *got.StartDate)
	}

	if !got.EndDate.Equal(now) {
		t.Errorf("expected end date %v, got %v", now, *got.EndDate)
	}
}

func playerNames(gains []wiseoldman.PlayerGains) []string {
	names := make([]string, len(gains))
	for i, gain := range gains {
		names[i] = gain.PlayerName
	}

	return names
}

func playerUsernames(players []Player) []string {
	names := make([]string, len(players))
	for i, player := range players {
		names[i] = player.Username
	}

	return names
}
