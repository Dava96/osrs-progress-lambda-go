package main

import (
	"context"
	"encoding/json"
	"fmt"
	"osrs-progress-lambda-go/config"
	"osrs-progress-lambda-go/discord"
	"osrs-progress-lambda-go/wiseoldman"
	stdsort "sort"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type Players struct {
	Players []Player
}

type Player struct {
	Username      string
	LastChangedAt *time.Time
	PlayerGains   wiseoldman.PlayerGains
	Err           error
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	wClient, err := wiseoldman.NewClient(wiseoldman.ApiUrl)
	if err != nil {
		return err
	}

	timeRange := resolveGainsPeriodFromConfig(&cfg, time.Now())

	players := updatePlayers(wClient, cfg.Usernames)

	filteredList, err := filter(players, timeRange.StartDate)
	if err != nil {
		return err
	}

	data := getPlayerData(wClient, filteredList, timeRange)
	sorted, err := sortPlayerGains(&cfg.SortBy, playerGainsFromPlayers(data.Players))
	if err != nil {
		return err
	}

	embed, err := discord.NewEmbed(sorted, discord.EmbedData{
		Colour:    cfg.EmbedColour,
		IconUrl:   cfg.ImageUrl,
		Thumbnail: cfg.ThumbnailUrl,
		Timezone:  cfg.Timezone.String(),
	}, cfg.SortBy, cfg.GainsPeriod)
	if err != nil {
		return err
	}

	dClient, err := discord.NewClient(cfg.WebhookUrl)
	if err != nil {
		return err
	}

	_, err = dClient.Send(embed)
	if err != nil {
		return err
	}

	for _, player := range data.Players {
		if player.Err != nil {
			fmt.Printf("player %s failed: %v\n", player.Username, player.Err)
			continue
		}
	}

	return nil
}

func playerGainsFromPlayers(players []Player) []wiseoldman.PlayerGains {
	pgs := make([]wiseoldman.PlayerGains, 0, len(players))

	for _, player := range players {
		if player.Err != nil {
			continue
		}

		if !hasAnyGains(player.PlayerGains) {
			continue
		}

		pgs = append(pgs, player.PlayerGains)
	}

	return pgs
}

func hasAnyGains(gains wiseoldman.PlayerGains) bool {
	exp := gains.PlayerData.Skills["overall"].Experience.Gained
	ehp := gains.PlayerData.Computed["ehp"].Value.Gained
	ehb := gains.PlayerData.Computed["ehb"].Value.Gained

	return exp > 0 || ehp > 0 || ehb > 0
}

func filter(players *Players, period *time.Time) (*Players, error) {
	if players == nil {
		return nil, fmt.Errorf("players is nil")
	}

	filtered := Players{
		Players: make([]Player, 0, len(players.Players)),
	}

	for _, player := range players.Players {
		if player.Err != nil {
			fmt.Printf("player %s failed: %v\n", player.Username, player.Err)
			continue
		}

		if period != nil {
			if player.LastChangedAt == nil {
				continue
			}

			if player.LastChangedAt.Before(*period) {
				continue
			}
		}

		filtered.Players = append(filtered.Players, player)
	}

	return &filtered, nil
}

func updatePlayers(client *wiseoldman.Client, userNames []string) *Players {
	completedPlayers := make(chan Player, len(userNames))
	concurrencyLimit := make(chan struct{}, 5)

	var waitForPlayers sync.WaitGroup

	for _, username := range userNames {
		waitForPlayers.Add(1)

		go func(username string) {
			defer waitForPlayers.Done()
			concurrencyLimit <- struct{}{}
			defer func() { <-concurrencyLimit }()

			playerUpdateResponse, err := client.SendUpdateForPlayer(username)

			player := Player{
				Username: username,
				Err:      err,
			}

			if err == nil {
				player.LastChangedAt = playerUpdateResponse.LastChangedAt
			}

			completedPlayers <- player
		}(username)
	}

	waitForPlayers.Wait()
	close(completedPlayers)
	var players Players
	for result := range completedPlayers {
		players.Players = append(players.Players, result)
	}

	return &players
}

func getPlayerData(client *wiseoldman.Client, players *Players, params *wiseoldman.PlayerGainsParams) Players {
	completedPlayers := make(chan Player, len(players.Players))
	concurrencyLimit := make(chan struct{}, 5)

	var waitForPlayers sync.WaitGroup

	for _, player := range players.Players {
		waitForPlayers.Add(1)

		go func(player Player) {
			defer waitForPlayers.Done()
			concurrencyLimit <- struct{}{}
			defer func() { <-concurrencyLimit }()

			playerDataResponse, err := client.GetPlayerData(
				player.Username,
				params,
			)

			if err != nil {
				player.Err = err
			} else {
				player.PlayerGains = playerDataResponse
			}

			completedPlayers <- player
		}(player)
	}
	waitForPlayers.Wait()
	close(completedPlayers)

	var updatedPlayers Players
	for player := range completedPlayers {
		updatedPlayers.Players = append(updatedPlayers.Players, player)
	}

	return updatedPlayers
}

func resolveGainsPeriodFromConfig(cfg *config.Config, now time.Time) *wiseoldman.PlayerGainsParams {
	if cfg.GainsQueryMode == config.GainsQueryModePeriod {
		return &wiseoldman.PlayerGainsParams{
			GainsPeriod: &cfg.GainsPeriod,
		}
	}

	end := now.In(cfg.Timezone)

	var start time.Time

	period := cfg.GainsPeriod
	switch period {
	case wiseoldman.GainsPeriodFiveMin:
		start = end.Add(-5 * time.Minute)
	case wiseoldman.GainsPeriodDay:
		start = end.AddDate(0, 0, -1)
	case wiseoldman.GainsPeriodWeek:
		start = end.AddDate(0, 0, -7)
	case wiseoldman.GainsPeriodMonth:
		start = end.AddDate(0, -1, 0)
	case wiseoldman.GainsPeriodYear:
		start = end.AddDate(-1, 0, 0)
	}

	return &wiseoldman.PlayerGainsParams{
		StartDate: &start,
		EndDate:   &end,
	}
}

func sortPlayerGains(sortBy *config.SortBy, gains []wiseoldman.PlayerGains) ([]wiseoldman.PlayerGains, error) {
	if sortBy == nil {
		return nil, fmt.Errorf("sort by is required")
	}

	sortedGains := make([]wiseoldman.PlayerGains, len(gains))
	copy(sortedGains, gains)

	stdsort.Slice(sortedGains, func(i, j int) bool {
		return sortValue(*sortBy, sortedGains[i]) > sortValue(*sortBy, sortedGains[j])
	})

	return sortedGains, nil
}

func sortValue(sortBy config.SortBy, gain wiseoldman.PlayerGains) float64 {
	switch sortBy {
	case config.SortByExp:
		return float64(gain.PlayerData.Skills["overall"].Experience.Gained)
	case config.SortByEhp:
		return gain.PlayerData.Computed["ehp"].Value.Gained
	case config.SortByEhb:
		return gain.PlayerData.Computed["ehb"].Value.Gained
	default:
		return 0
	}
}
