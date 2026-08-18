package discord

import (
	"fmt"
	"net/url"
	"osrs-progress-lambda-go/config"
	"osrs-progress-lambda-go/wiseoldman"
	"strconv"
	"strings"
)

const defaultOrangeDecimalColour = 16744805

type Embed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Colour      int       `json:"color"`
	Fields      []Field   `json:"fields"`
	Author      Author    `json:"author"`
	Footer      Footer    `json:"footer"`
	Thumbnail   Thumbnail `json:"thumbnail"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Author struct {
	Name    string `json:"name"`
	IconUrl string `json:"icon_url"`
}

type Footer struct {
	Text string `json:"text"`
}

type Thumbnail struct {
	ThumbnailUrl string `json:"url"`
}

type EmbedData struct {
	Colour    string
	IconUrl   string
	Thumbnail string
	Timezone  string
}

func NewEmbed(playerGains []wiseoldman.PlayerGains, data EmbedData, by config.SortBy, period wiseoldman.GainsPeriod) (Embed, error) {
	colourValue, err := parseColourToDecimal(data.Colour)
	if err != nil {
		return Embed{}, err
	}

	icon, err := parseImageUrl(data.IconUrl)
	if err != nil {
		return Embed{}, err
	}

	thumb, err := parseImageUrl(data.Thumbnail)
	if err != nil {
		return Embed{}, err
	}

	return Embed{
		Title:       "OSRS Activity",
		Description: fmt.Sprintf("Ranks Players in the group by %s over the past %s", strings.ToTitle(string(by)), strings.ToTitle(string(period))),
		Colour:      colourValue,
		Thumbnail: Thumbnail{
			ThumbnailUrl: thumb.String(),
		},
		Author: Author{
			IconUrl: icon.String(),
			Name:    "OSRS Activity",
		},
		Footer: Footer{
			Text: fmt.Sprintf("Timezone - %s", data.Timezone),
		},
		Fields: buildFieldsFromPlayerData(playerGains),
	}, nil
}

func parseColourToDecimal(colour string) (int, error) {
	if colour == "" {
		return defaultOrangeDecimalColour, nil
	}

	colour = strings.TrimSpace(colour)
	colour = strings.TrimPrefix(colour, "#")
	colour = strings.TrimPrefix(colour, "0x")
	colour = strings.TrimPrefix(colour, "0X")

	decimal, err := strconv.ParseInt(colour, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse colour %v", colour)
	}

	return int(decimal), nil
}

func parseImageUrl(imageUrl string) (*url.URL, error) {
	imgUrl, err := url.Parse(imageUrl)
	if err != nil {
		return nil, fmt.Errorf("could not parse thumbnail url %s: %w", imgUrl, err)
	}

	return imgUrl, nil
}

func buildFieldsFromPlayerData(playerGains []wiseoldman.PlayerGains) (fields []Field) {
	fields = make([]Field, len(playerGains))
	for index, player := range playerGains {
		fields[index] = Field{
			player.PlayerName,
			fmt.Sprintf(
				"**EXP:** `%s`\t**EHP:** `%.3f`\t**EHB:** `%.3f`",
				formatInt(player.PlayerData.Skills["overall"].Experience.Gained),
				player.PlayerData.Computed["ehp"].Value.Gained,
				player.PlayerData.Computed["ehb"].Value.Gained,
			),
			false,
		}
	}

	return fields
}

func formatInt(number int) string {
	s := strconv.Itoa(number)

	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
