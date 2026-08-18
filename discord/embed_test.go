package discord

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"osrs-progress-lambda-go/config"
	"osrs-progress-lambda-go/wiseoldman"
	"testing"
)

func TestFormatInts(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{
			name:  "100",
			input: 100,
			want:  "100",
		},
		{
			name:  "1000",
			input: 1000,
			want:  "1,000",
		},
		{
			name:  "10000",
			input: 10000,
			want:  "10,000",
		},
		{
			name:  "100000",
			input: 100000,
			want:  "100,000",
		},
		{
			name:  "10000000",
			input: 1000000,
			want:  "1,000,000",
		},
		{
			name:  "100000000",
			input: 10000000,
			want:  "10,000,000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if formatInt(test.input) != test.want {
				t.Errorf("got %q, want %q", formatInt(test.input), test.want)
			}
		})
	}
}

func TestNewEmbed(t *testing.T) {
	const expectedEXP = 1_000_000
	const expectedEHP = 126.26093
	const expectedEHB = 127.26093

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

	mockEmbed := mockEmbeds.Embeds[0]

	embedData := EmbedData{
		Colour:    "",
		IconUrl:   "https://oldschool.runescape.wiki/images/Cheer_%28Penguin%29_emote_icon.png?e60bd",
		Thumbnail: "https://oldschool.runescape.wiki/images/Cheer_%28Penguin%29_emote_icon.png?e60bd",
		Timezone:  "London/Europe",
	}

	mockGains := make([]wiseoldman.PlayerGains, 4)

	for index := range mockGains {

		gains := getPlayerGainsResponse(t)

		gains.PlayerName = fmt.Sprintf("MyRsn%d", index)

		overall := gains.PlayerData.Skills["overall"]
		overall.Experience.Gained = expectedEXP
		gains.PlayerData.Skills["overall"] = overall

		ehp := gains.PlayerData.Computed["ehp"]
		ehp.Value.Gained = expectedEHP
		gains.PlayerData.Computed["ehp"] = ehp

		ehb := gains.PlayerData.Computed["ehb"]
		ehb.Value.Gained = expectedEHB
		gains.PlayerData.Computed["ehb"] = ehb

		mockGains[index] = gains
	}

	embed, err := NewEmbed(mockGains, embedData, config.SortByEhp, wiseoldman.GainsPeriodDay)
	if err != nil {
		t.Errorf("error creating embed: %s", err)
	}

	if embed.Colour != defaultOrangeDecimalColour {
		t.Errorf("expected colour to be %d, got %d", defaultOrangeDecimalColour, embed.Colour)
	}

	if embed.Author.Name != mockEmbed.Author.Name {
		t.Errorf("expected author name to be %s, got %s", mockEmbed.Author.Name, embed.Author.Name)
	}

	expectedDescription := "Ranks Players in the group by EHP over the past DAY"
	if embed.Description != expectedDescription {
		t.Errorf("expected description to be %s, got %s", expectedDescription, embed.Description)
	}

	for i := range mockEmbed.Fields {
		if embed.Fields[i].Inline != mockEmbed.Fields[i].Inline {
			t.Errorf("expected inline to be %v, got %v", mockEmbed.Fields[i].Inline, embed.Fields[i].Inline)
		}

		if embed.Fields[i].Name != mockEmbed.Fields[i].Name {
			t.Errorf("expected field name to be %s, got %s", embed.Fields[i].Name, mockEmbed.Fields[i].Name)
		}

		if embed.Fields[i].Value != mockEmbed.Fields[i].Value {
			t.Errorf("expected field value to be %s, got %s", embed.Fields[i].Value, mockEmbed.Fields[i].Value)
		}
	}
}

func getPlayerGainsResponse(t *testing.T) wiseoldman.PlayerGains {
	t.Helper()
	body, err := os.ReadFile("testdata/player_gains.json")

	if err != nil {
		t.Fatalf("error reading client response %v", err)
	}

	var gains wiseoldman.PlayerGains
	if err := json.Unmarshal(body, &gains); err != nil {
		t.Fatalf("error unmarshaling client response %v", err)
	}

	return gains
}

func TestParseColourToDecimal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid colour",
			input: "#eb4034",
			want:  15417396,
		},
		{
			name:  "fall back to default colour",
			input: "",
			want:  defaultOrangeDecimalColour,
		},
		{
			name:    "error",
			input:   "not a real colour",
			want:    defaultOrangeDecimalColour,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseColourToDecimal(test.input)

			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil && !test.wantErr {
				t.Errorf("got error %v", err)
			}

			if got != test.want {
				t.Errorf("got %d, want %d", got, test.want)
			}

		})
	}
}

func TestParseImageUrl(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *url.URL
		wantErr bool
	}{
		{
			name:  "valid url",
			input: "https://oldschool.runescape.wiki/images/Cheer_%28Penguin%29_emote_icon.png?e60bd",
			want: new(url.URL{
				Scheme:     "https",
				Opaque:     "",
				User:       nil,
				Host:       "oldschool.runescape.wiki",
				Path:       "/images/Cheer_(Penguin)_emote_icon.png",
				Fragment:   "",
				RawQuery:   "e60bd",
				ForceQuery: false,
				OmitHost:   false,
			}),
		},
		{
			name:  "empty url",
			input: "",
			want: new(url.URL{
				Scheme:     "",
				Opaque:     "",
				User:       nil,
				Host:       "",
				Path:       "",
				Fragment:   "",
				RawQuery:   "",
				ForceQuery: false,
				OmitHost:   false,
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseImageUrl(test.input)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if test.want.String() != got.String() {
				t.Errorf("got %s, want %s", got.Path, test.want.Path)
			}
		})
	}
}

func TestAddEmojiForGainsPosition(t *testing.T) {
	type args struct {
		position int
		players  int
	}
	tests := []struct {
		name  string
		input args
		want  string
	}{
		{
			name: "first position",
			input: args{
				position: 0,
				players:  1,
			},
			want: ":first_place:",
		},
		{
			name: "second position",
			input: args{
				position: 1,
				players:  4,
			},
			want: ":second_place:",
		},
		{
			name: "third position",
			input: args{
				position: 2,
				players:  4,
			},
			want: ":third_place:",
		},
		{
			name: "last position",
			input: args{
				position: 3,
				players:  4,
			},
			want: ":poop:",
		},
		{
			name: "fourth position",
			input: args{
				position: 4,
				players:  6,
			},
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := addEmojiForGainsPosition(test.input.position, test.input.players)

			if got != test.want {
				t.Errorf("got %s, expected %s", got, test.want)
			}
		})
	}
}
