package wiseoldman

import (
	"fmt"
	"net/url"
	"time"
)

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type RateLimitHeaders struct {
	RateLimit          int
	RateLimitRemaining int
	RateLimitReset     int
}

type PlayerUpdate struct {
	RateLimitHeaders RateLimitHeaders
	LastChangedAt    *time.Time `json:"lastChangedAt"`
}

type PlayerGains struct {
	RateLimitHeaders RateLimitHeaders
	PlayerName       string     `json:"playerName"`
	StartsAt         *time.Time `json:"startsAt"`
	EndsAt           *time.Time `json:"endsAt"`
	PlayerData       PlayerData `json:"data"`
}

type PlayerData struct {
	Skills     map[string]Skill          `json:"skills"`
	Bosses     map[string]Boss           `json:"bosses"`
	Activities map[string]Activity       `json:"activities"`
	Computed   map[string]ComputedMetric `json:"computed"`
}

type Skill struct {
	SkillName  string            `json:"metric"`
	Experience SharedIntMetric   `json:"experience"`
	Ehp        SharedFloatMetric `json:"ehp"`
	Rank       SharedIntMetric   `json:"rank"`
	Level      SharedIntMetric   `json:"level"`
}

type Boss struct {
	BossName string            `json:"metric"`
	Ehb      SharedFloatMetric `json:"ehb"`
	Rank     SharedIntMetric   `json:"rank"`
	Kills    SharedIntMetric   `json:"kills"`
}

type Activity struct {
	ActivityName string          `json:"metric"`
	Rank         SharedIntMetric `json:"rank"`
	Score        SharedIntMetric `json:"score"`
}

type ComputedMetric struct {
	Name  string            `json:"metric"`
	Value SharedFloatMetric `json:"value"`
	Rank  SharedIntMetric   `json:"rank"`
}

// SharedIntMetric This metric is shared between Rank, Level and Experience
type SharedIntMetric struct {
	Gained    int `json:"gained"`
	StartedAt int `json:"start"`
	EndedAt   int `json:"end"`
}

// SharedFloatMetric is shared between EHP / EHB
type SharedFloatMetric struct {
	Gained    float64 `json:"gained"`
	StartedAt float64 `json:"start"`
	EndedAt   float64 `json:"end"`
}

type GainsPeriod string

const (
	GainsPeriodFiveMin GainsPeriod = "five_min"
	GainsPeriodDay     GainsPeriod = "day"
	GainsPeriodWeek    GainsPeriod = "week"
	GainsPeriodMonth   GainsPeriod = "month"
	GainsPeriodYear    GainsPeriod = "year"
)

type PlayerGainsParams struct {
	GainsPeriod *GainsPeriod
	StartDate   *time.Time
	EndDate     *time.Time
}

func (p PlayerGainsParams) Validate() error {
	if p.GainsPeriod == nil && p.StartDate == nil && p.EndDate == nil {
		return fmt.Errorf("please set at least one query parameter: period, startDate, endDate")
	}

	if p.StartDate != nil && p.EndDate != nil {
		if p.EndDate.Before(*p.StartDate) {
			return fmt.Errorf("please set the end date to be after the start date")
		}
	}

	if p.EndDate == nil && p.StartDate != nil || p.EndDate != nil && p.StartDate == nil {
		return fmt.Errorf("please set both start date and end date")
	}

	return nil
}

const DateLayout = "2006-01-02T15:04:05.000Z"

// GainsPeriod takes priority over StartDate and EndDate
// You either want GainsPeriod or StartDate and EndDate not both
func (p PlayerGainsParams) QueryValues() url.Values {
	params := url.Values{}
	if p.GainsPeriod != nil && p.StartDate == nil && p.EndDate == nil {
		params.Set("period", string(*p.GainsPeriod))
	}

	if p.StartDate != nil && p.EndDate != nil && p.GainsPeriod == nil {
		params.Set("startDate", p.StartDate.UTC().Format(DateLayout))
		params.Set("endDate", p.EndDate.UTC().Format(DateLayout))
	}

	return params
}
