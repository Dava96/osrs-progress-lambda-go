package config

import (
	"errors"
	"fmt"
	"os"
	"osrs-progress-lambda-go/wiseoldman"
	"strings"
	"time"
)

const MaximumCharacterNameLength = 12

type SortBy string

const (
	SortByExp SortBy = "exp"
	SortByEhp SortBy = "ehp"
	SortByEhb SortBy = "ehb"
)

type GainsQueryMode string

const (
	GainsQueryModePeriod GainsQueryMode = "period"
	GainsQueryModeRange  GainsQueryMode = "range"
)

type Config struct {
	Usernames      []string
	SortBy         SortBy
	WebhookUrl     string
	GainsQueryMode GainsQueryMode
	GainsPeriod    wiseoldman.GainsPeriod
	Timezone       *time.Location
	ImageUrl       string
	ThumbnailUrl   string
	EmbedColour    string
}

func LoadFromEnv() (Config, error) {
	var config Config
	var err error

	usernames, err := requiredEnv("USERNAMES")
	if err != nil {
		return Config{}, err
	}

	config.Usernames, err = parseUsernames(usernames)
	if err != nil {
		return Config{}, err
	}

	sortBy, err := requiredEnv("SORT_BY")
	if err != nil {
		return Config{}, err
	}

	config.SortBy, err = parseSortBy(sortBy)
	if err != nil {
		return Config{}, err
	}

	config.WebhookUrl, err = requiredEnv("WEBHOOK_URL")
	if err != nil {
		return Config{}, err
	}

	gainsQueryMode, err := requiredEnv("GAINS_QUERY_MODE")
	if err != nil {
		return Config{}, err
	}

	config.GainsQueryMode, err = parseGainsQueryMode(gainsQueryMode)
	if err != nil {
		return Config{}, err
	}

	gainsPeriod, err := requiredEnv("GAINS_PERIOD")
	if err != nil {
		return Config{}, err
	}

	config.GainsPeriod, err = parseGainsPeriod(gainsPeriod)
	if err != nil {
		return Config{}, err
	}

	timezone := os.Getenv("TIMEZONE")
	config.Timezone, err = parseTimezone(timezone)
	if err != nil {
		return Config{}, err
	}

	config.ImageUrl = os.Getenv("IMAGE_URL")
	config.EmbedColour = os.Getenv("EMBED_COLOUR")
	config.ThumbnailUrl = os.Getenv("THUMBNAIL_URL")

	return config, nil
}

func (c *Config) Validate() error {
	if len(c.Usernames) == 0 {
		return errors.New("no usernames specified")
	}

	if c.SortBy == "" {
		return errors.New("sort by is required")
	}

	if c.WebhookUrl == "" {
		return errors.New("webhook url is required")
	}

	return nil
}

func parseUsernames(usernames string) ([]string, error) {
	if len(usernames) == 0 {
		return []string{}, errors.New("no usernames specified")
	}

	if len(usernames) > MaximumCharacterNameLength && !strings.Contains(usernames, ",") {
		return []string{}, fmt.Errorf("list of usernames exceeds maximum character name length of: %d and does not contain a comma", MaximumCharacterNameLength)
	}

	if !strings.Contains(usernames, ",") {
		return []string{}, errors.New("usernames need to be separated with a comma")
	}

	parts := strings.Split(usernames, ",")
	names := make([]string, 0, len(parts))

	for _, name := range parts {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no usernames specified")
	}

	return removeDuplicateUsernames(names), nil
}

func parseSortBy(value string) (SortBy, error) {
	sortBy := SortBy(strings.ToLower(strings.TrimSpace(value)))

	switch sortBy {
	case SortByExp, SortByEhp, SortByEhb:
		return sortBy, nil
	default:
		return "", fmt.Errorf("invalid SORT_BY %q, must be one of {exp, ehp, ehb}", sortBy)
	}
}

func parseGainsQueryMode(value string) (GainsQueryMode, error) {
	mode := GainsQueryMode(strings.ToLower(strings.TrimSpace(value)))

	switch mode {
	case GainsQueryModePeriod, GainsQueryModeRange:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid GAINS_QUERY_MODE %q, must be one of {period, range}", mode)
	}
}

func requiredEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)

	if !ok || strings.TrimSpace(value) == "" {
		return value, fmt.Errorf("environment variable %s is required", name)
	}

	return value, nil
}

func parseGainsPeriod(value string) (wiseoldman.GainsPeriod, error) {
	gainsPeriod := wiseoldman.GainsPeriod(strings.ToLower(strings.TrimSpace(value)))

	switch gainsPeriod {
	case "5min", "five_min", "5m":
		return wiseoldman.GainsPeriodFiveMin, nil
	case "day", "1d", "1day":
		return wiseoldman.GainsPeriodDay, nil
	case "week", "1week", "1 week", "1_week":
		return wiseoldman.GainsPeriodWeek, nil
	case "1month", "1 month", "1_month", "month":
		return wiseoldman.GainsPeriodMonth, nil
	case "1year", "year", "1_year", "1 year":
		return wiseoldman.GainsPeriodYear, nil
	default:
		return "", fmt.Errorf("invalid gains period: %s", value)
	}
}

func parseTimezone(value string) (*time.Location, error) {
	location, err := time.LoadLocation(value)

	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %s", value)
	}

	if value == "" {
		fmt.Println("TIMEZONE is empty and will be set to UTC")
	}

	return location, nil
}

func removeDuplicateUsernames(usernames []string) []string {
	names := make([]string, 0, len(usernames))
	seen := make(map[string]bool)

	for _, username := range usernames {
		key := strings.ToLower(username)

		if seen[key] {
			continue
		}

		seen[key] = true
		names = append(names, username)
	}

	return names
}
