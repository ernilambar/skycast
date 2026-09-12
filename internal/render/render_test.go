package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ernilambar/skycast/internal/weather"
)

func TestCodeToConditionMapsWMOCodes(t *testing.T) {
	cases := []struct {
		name string
		code int
		want string
	}{
		{"clear sky", 0, "Clear"},
		{"mainly clear", 1, "Clear"},
		{"partly cloudy", 2, "Clouds"},
		{"overcast", 3, "Clouds"},
		{"fog", 45, "Fog"},
		{"depositing rime fog", 48, "Fog"},
		{"light drizzle", 51, "Rain"},
		{"moderate drizzle", 53, "Rain"},
		{"dense drizzle", 55, "Rain"},
		{"light freezing drizzle", 56, "Rain"},
		{"dense freezing drizzle", 57, "Rain"},
		{"slight rain", 61, "Rain"},
		{"moderate rain", 63, "Rain"},
		{"heavy rain", 65, "Rain"},
		{"light freezing rain", 66, "Rain"},
		{"heavy freezing rain", 67, "Rain"},
		{"slight snow", 71, "Snow"},
		{"moderate snow", 73, "Snow"},
		{"heavy snow", 75, "Snow"},
		{"snow grains", 77, "Snow"},
		{"slight rain showers", 80, "Rain"},
		{"moderate rain showers", 81, "Rain"},
		{"violent rain showers", 82, "Rain"},
		{"slight snow showers", 85, "Snow"},
		{"heavy snow showers", 86, "Snow"},
		{"thunderstorm", 95, "Thunderstorm"},
		{"thunderstorm with slight hail", 96, "Thunderstorm"},
		{"thunderstorm with heavy hail", 99, "Thunderstorm"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CodeToCondition(tc.code); got != tc.want {
				t.Errorf("CodeToCondition(%d) = %q, want %q", tc.code, got, tc.want)
			}
		})
	}
}

func TestCodeToConditionFallsBackToCloudsForUnknownCodes(t *testing.T) {
	for _, code := range []int{-1, 4, 30, 47, 50, 60, 70, 90} {
		if got := CodeToCondition(code); got != "Clouds" {
			t.Errorf("CodeToCondition(%d) = %q, want %q", code, got, "Clouds")
		}
	}
}

func TestUnitLabelsFollowUnits(t *testing.T) {
	if got := unitLabel("metric"); got != "°C" {
		t.Errorf("unitLabel(metric) = %q, want %q", got, "°C")
	}
	if got := unitLabel("imperial"); got != "°F" {
		t.Errorf("unitLabel(imperial) = %q, want %q", got, "°F")
	}
	if got := windUnitLabel("metric"); got != "km/h" {
		t.Errorf("windUnitLabel(metric) = %q, want %q", got, "km/h")
	}
	if got := windUnitLabel("imperial"); got != "mph" {
		t.Errorf("windUnitLabel(imperial) = %q, want %q", got, "mph")
	}
	if got := precipitationUnitLabel("metric"); got != "mm" {
		t.Errorf("precipitationUnitLabel(metric) = %q, want %q", got, "mm")
	}
	if got := precipitationUnitLabel("imperial"); got != "in" {
		t.Errorf("precipitationUnitLabel(imperial) = %q, want %q", got, "in")
	}
}

func TestRoundRoundsToNearestInt(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{0, 0},
		{0.4, 0},
		{0.5, 1},
		{-0.5, -1},
		{1.5, 2},
		{2.5, 3},
		{20.4, 20},
		{20.6, 21},
	}

	for _, tc := range cases {
		if got := round(tc.in); got != tc.want {
			t.Errorf("round(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestNumStrOmitsTrailingZero(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{5, "5"},
		{0, "0"},
		{60, "60"},
		{-3, "-3"},
		{1.2, "1.2"},
		{1.25, "1.25"},
	}

	for _, tc := range cases {
		if got := numStr(tc.in); got != tc.want {
			t.Errorf("numStr(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatTempShowsUnitAndThresholdEmoji(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		units string
		unit  string
		emoji string
	}{
		{"freezing metric", 0, "metric", "°C", "❄️"},
		{"just above freezing", 1, "metric", "°C", "🌤️"},
		{"cold metric", 10, "metric", "°C", "🌤️"},
		{"upper cold boundary", 18, "metric", "°C", "🌤️"},
		{"just above cold boundary", 19, "metric", "°C", "☀️"},
		{"upper mild boundary", 28, "metric", "°C", "☀️"},
		{"hot metric", 29, "metric", "°C", "🔥"},
		{"freezing imperial", 32, "imperial", "°F", "❄️"},
		{"mild imperial", 68, "imperial", "°F", "☀️"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatTemp(tc.value, tc.units)
			if !strings.Contains(got, tc.unit) {
				t.Errorf("formatTemp(%v, %s) = %q, want it to contain %q", tc.value, tc.units, got, tc.unit)
			}
			if !strings.Contains(got, tc.emoji) {
				t.Errorf("formatTemp(%v, %s) = %q, want it to contain %q", tc.value, tc.units, got, tc.emoji)
			}
		})
	}
}

func TestFormatTempConvertsImperialValue(t *testing.T) {
	if got := formatTemp(68, "imperial"); !strings.Contains(got, "68°F") {
		t.Errorf("formatTemp(68, imperial) = %q, want it to contain %q", got, "68°F")
	}
}

func TestLocationLabel(t *testing.T) {
	withCountry := weather.Location{City: "Kathmandu", Country: "Nepal"}
	if got := locationLabel(withCountry); got != "Kathmandu, Nepal" {
		t.Errorf("locationLabel(%+v) = %q, want %q", withCountry, got, "Kathmandu, Nepal")
	}

	withoutCountry := weather.Location{City: "Kathmandu"}
	if got := locationLabel(withoutCountry); got != "Kathmandu" {
		t.Errorf("locationLabel(%+v) = %q, want %q", withoutCountry, got, "Kathmandu")
	}
}

func TestWeekdayName(t *testing.T) {
	if got := weekdayName("2026-08-01"); got != "Sat" {
		t.Errorf("weekdayName(2026-08-01) = %q, want %q", got, "Sat")
	}
	if got := weekdayName("not-a-date"); got != "not-a-date" {
		t.Errorf("weekdayName(not-a-date) = %q, want the original string back", got)
	}
}

func TestFormatClock(t *testing.T) {
	if got := formatClock("2026-08-01T06:30"); got != "06:30" {
		t.Errorf("formatClock(2026-08-01T06:30) = %q, want %q", got, "06:30")
	}
	if got := formatClock(""); got != "--:--" {
		t.Errorf("formatClock(\"\") = %q, want %q", got, "--:--")
	}
	if got := formatClock("garbage"); got != "--:--" {
		t.Errorf("formatClock(garbage) = %q, want %q", got, "--:--")
	}
}

func TestTimeOfDayBucketClassifiesRelativeToSunriseAndSunset(t *testing.T) {
	const sunrise = "2026-08-01T06:00"
	const sunset = "2026-08-01T18:00"

	cases := []struct {
		name    string
		current string
		want    string
	}{
		{"midnight", "2026-08-01T00:00", "Night"},
		{"dawn start", "2026-08-01T05:30", "Dawn"},
		{"dawn", "2026-08-01T06:00", "Dawn"},
		{"dawn end", "2026-08-01T06:30", "Day"},
		{"day", "2026-08-01T12:00", "Day"},
		{"dusk start", "2026-08-01T17:30", "Dusk"},
		{"dusk", "2026-08-01T18:00", "Dusk"},
		{"dusk end", "2026-08-01T18:30", "Night"},
		{"night", "2026-08-01T22:00", "Night"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := timeOfDayBucket(tc.current, sunrise, sunset); got != tc.want {
				t.Errorf("timeOfDayBucket(%q, %q, %q) = %q, want %q", tc.current, sunrise, sunset, got, tc.want)
			}
		})
	}
}

func TestTimeOfDayBucketFallsBackToDayOnInvalidTimestamps(t *testing.T) {
	cases := []struct {
		name    string
		current string
		sunrise string
		sunset  string
	}{
		{"invalid current", "nope", "2026-08-01T06:00", "2026-08-01T18:00"},
		{"invalid sunrise", "2026-08-01T12:00", "nope", "2026-08-01T18:00"},
		{"invalid sunset", "2026-08-01T12:00", "2026-08-01T06:00", "nope"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := timeOfDayBucket(tc.current, tc.sunrise, tc.sunset); got != "Day" {
				t.Errorf("timeOfDayBucket(%q, %q, %q) = %q, want %q", tc.current, tc.sunrise, tc.sunset, got, "Day")
			}
		})
	}
}

func sampleCurrent() weather.CurrentWeather {
	return weather.CurrentWeather{
		Temperature:   21,
		FeelsLike:     23,
		Humidity:      60,
		Wind:          5,
		Precipitation: 1.2,
		Code:          3,
		CloudCover:    40,
		Time:          "2026-08-01T14:30",
		Sunrise:       "2026-08-01T05:45",
		Sunset:        "2026-08-01T19:10",
	}
}

func TestCurrentWeatherPlainOutput(t *testing.T) {
	var buf bytes.Buffer

	location := weather.Location{City: "Kathmandu", Country: "Nepal"}
	CurrentWeather(&buf, location, sampleCurrent(), "metric", true)

	out := buf.String()

	for _, want := range []string{
		"Location: Kathmandu, Nepal",
		"Condition: Clouds",
		"Temperature: 21°C",
		"Feels Like: 23°C",
		"Humidity: 60%",
		"Wind Speed: 5 km/h",
		"Precipitation: 1.2 mm",
		"Cloud Cover: 40%",
		"Time of Day: Day",
		"Sunrise: 05:45",
		"Sunset: 19:10",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plain output = %q, want it to contain %q", out, want)
		}
	}

	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain output should not contain ANSI codes: %q", out)
	}
}

func TestCurrentWeatherPlainOutputUsesImperialUnits(t *testing.T) {
	var buf bytes.Buffer

	current := sampleCurrent()
	current.Temperature = 70
	current.Wind = 9
	current.Precipitation = 0.5

	CurrentWeather(&buf, weather.Location{City: "Tokyo"}, current, "imperial", true)

	out := buf.String()
	for _, want := range []string{
		"Temperature: 70°F",
		"Wind Speed: 9 mph",
		"Precipitation: 0.5 in",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plain output = %q, want it to contain %q", out, want)
		}
	}
}

func TestCurrentWeatherCardOutput(t *testing.T) {
	var buf bytes.Buffer

	location := weather.Location{City: "Kathmandu", Country: "Nepal"}
	CurrentWeather(&buf, location, sampleCurrent(), "metric", false)

	out := buf.String()
	for _, want := range []string{
		"📍 KATHMANDU, NEPAL",
		"Condition:",
		"Clouds",
		"Temperature:",
		"Time of Day:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("card output = %q, want it to contain %q", out, want)
		}
	}

	if !strings.Contains(out, "\x1b[") {
		t.Errorf("card output should contain ANSI codes: %q", out)
	}
}

func sampleForecast() []weather.DailyForecast {
	return []weather.DailyForecast{
		{Date: "2026-08-01", Code: 3, TempMax: 30, TempMin: 20, PrecipitationProbability: 10},
		{Date: "2026-08-02", Code: 61, TempMax: 28, TempMin: 19, PrecipitationProbability: 55},
	}
}

func TestForecastPlainOutput(t *testing.T) {
	var buf bytes.Buffer

	Forecast(&buf, sampleForecast(), "metric", true)

	out := buf.String()
	for _, want := range []string{
		"5-Day Forecast:",
		"Sat: Clouds, High 30°C, Low 20°C, Precipitation 10%",
		"Sun: Rain, High 28°C, Low 19°C, Precipitation 55%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plain output = %q, want it to contain %q", out, want)
		}
	}

	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain output should not contain ANSI codes: %q", out)
	}
}

func TestForecastCardOutput(t *testing.T) {
	var buf bytes.Buffer

	Forecast(&buf, sampleForecast(), "metric", false)

	out := buf.String()
	for _, want := range []string{
		"5-Day Forecast",
		"Sat",
		"Sun",
		"Clouds",
		"Rain",
		"☁️",
		"🌧️",
		"30°",
		"💧",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("card output = %q, want it to contain %q", out, want)
		}
	}

	if !strings.Contains(out, "\x1b[") {
		t.Errorf("card output should contain ANSI codes: %q", out)
	}
}
