package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/olexsmir/viye/internal/config"
	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string               { return "weather" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "weather" }
func (Tool) Execute(c *viye.Context) (string, error) {
	city, err := getCity(c)
	if err != nil {
		return "", err
	}

	fc, err := forecast(city)
	if err != nil {
		return "", err
	}

	return viye.FormatOutput(render(city, fc)), nil
}

func getCity(c *viye.Context) (string, error) {
	if len(c.Args) == 1 {
		return c.Args[0], nil
	}
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	city, ok := cfg.Get("city")
	if !ok {
		return "", errors.New("please provide a city option")
	}
	return city, nil
}

func render(city string, fc weatherResp) []string {
	if len(fc.Weather) < 3 || len(fc.Weather[0].Hourly) < 8 {
		return []string{"no forecast data"}
	}

	t := fc.Weather[0].Hourly
	lines := []string{}
	lines = append(lines, fmt.Sprintf("%s (%s)", city, fc.Weather[0].Date))
	lines = append(lines, fmt.Sprintf("Today %s %s° / %s %s° / %s %s°",
		emoji(t[2].Code), t[2].TempC, // morning
		emoji(t[4].Code), t[4].TempC, // midday
		emoji(t[7].Code), t[7].TempC)) // evening

	for _, d := range fc.Weather[1:3] {
		mi, _ := strconv.Atoi(d.Min)
		ma, _ := strconv.Atoi(d.Max)
		lines = append(lines, fmt.Sprintf("%-5s %s %.0f° (avg)",
			weekday(d.Date), emoji(d.Hourly[4].Code), (float64(mi)+float64(ma))/2))
	}
	return lines
}

func weekday(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Weekday().String()[:3]
}

// api

type weatherResp struct {
	Weather []day `json:"weather"`
}

// hour is one 3-hourly reading; 8 readings cover a day, hours 0..21.
type hour struct {
	TempC string `json:"tempC"`
	Code  string `json:"weatherCode"`
}

type day struct {
	Date   string `json:"date"`
	Min    string `json:"mintempC"`
	Max    string `json:"maxtempC"`
	Hourly []hour `json:"hourly"`
}

// forecast fetches the 3-day forecast for a city via wttr.in (Open-Meteo data).
func forecast(city string) (weatherResp, error) {
	u := "https://wttr.in/" + url.QueryEscape(city) + "?format=j1"
	ctx, cancel := context.WithTimeout(context.Background(), viye.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return weatherResp{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return weatherResp{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return weatherResp{}, fmt.Errorf("api returned %s", resp.Status)
	}

	var out weatherResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return weatherResp{}, err
	}
	if len(out.Weather) == 0 {
		return weatherResp{}, fmt.Errorf("city %q not found", city)
	}
	return out, nil
}

// emoji maps a wttr.in weather code to its phenomenon emoji.
func emoji(code string) string {
	c, _ := strconv.Atoi(code)
	switch {
	case c == 113:
		return "☀️"
	case c == 116:
		return "🌤️"
	case c == 119 || c == 122:
		return "☁️"
	case c == 143 || c == 248 || c == 260:
		return "🌫️"
	case c == 179 || c == 227 || c == 230 ||
		c >= 317 && c <= 338 || c == 350 ||
		c >= 362 && c <= 377 || c == 392 || c == 395: // snow, sleet
		return "🌨️"
	case c == 176 || c == 182 || c == 185 || c == 200 ||
		c >= 263 && c <= 314 || c >= 353 && c <= 359 ||
		c == 386 || c == 389: // drizzle, rain
		return "🌧️"
	default:
		return "🌡️"
	}
}
