package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ernilambar/skycast/internal/spinner"
	"github.com/ernilambar/skycast/internal/weather"
)

// stubSpinner routes spinner output into a buffer and speeds up its refresh
// interval so tests are fast and quiet.
func stubSpinner(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	original := newSpinner
	t.Cleanup(func() { newSpinner = original })

	newSpinner = func() *spinner.Spinner {
		sp := spinner.New()
		sp.SetOutput(buf)
		sp.SetInterval(time.Millisecond)
		return sp
	}

	return buf
}

// preserveWeatherSeams snapshots the weather function seams and restores them
// when the test finishes.
func preserveWeatherSeams(t *testing.T) {
	t.Helper()

	geocode, locate, fetch := geocodeCity, locateByIP, fetchWeather
	t.Cleanup(func() {
		geocodeCity, locateByIP, fetchWeather = geocode, locate, fetch
	})
}

func TestNewRootCmdRejectsInvalidUnits(t *testing.T) {
	cmd := newRootCmd("test")

	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--units", "kelvin", "Tokyo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want an invalid units error")
	}
	if !strings.Contains(err.Error(), "invalid units") {
		t.Errorf("Execute() error = %v, want it to mention \"invalid units\"", err)
	}
	if !strings.Contains(errOut.String(), "invalid units") {
		t.Errorf("stderr = %q, want it to report the invalid units error", errOut.String())
	}
}

func TestNewRootCmdRejectsTooManyArgs(t *testing.T) {
	cmd := newRootCmd("test")
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"Tokyo", "Paris"})

	if err := cmd.Execute(); err == nil {
		t.Error("Execute() error = nil, want an error for too many arguments")
	}
}

func TestNewRootCmdPrintsVersion(t *testing.T) {
	cmd := newRootCmd("1.2.3")

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "1.2.3") {
		t.Errorf("output = %q, want it to contain the version", out.String())
	}
}

func TestPrintErrorWritesRedMessage(t *testing.T) {
	var buf bytes.Buffer

	printError(&buf, errors.New("boom"))

	got := buf.String()
	if !strings.Contains(got, "Error: boom") {
		t.Errorf("printError() = %q, want it to contain the message", got)
	}
	if !strings.Contains(got, "\x1b[31m") {
		t.Errorf("printError() = %q, want it to color the message red", got)
	}
}

func TestRunRendersWeatherForCity(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(city string) (weather.Location, error) {
		if city != "Tokyo" {
			t.Errorf("GeocodeCity() city = %q, want %q", city, "Tokyo")
		}
		return weather.Location{City: "Tokyo", Country: "Japan", Latitude: 35.6, Longitude: 139.7}, nil
	}
	locateByIP = func() (weather.Location, error) {
		t.Error("LocateByIP() should not be called when a city is given")
		return weather.Location{}, nil
	}
	fetchWeather = func(latitude, longitude float64, units string, forecast bool) (weather.Weather, error) {
		if latitude != 35.6 || longitude != 139.7 {
			t.Errorf("FetchWeather() coordinates = (%v, %v), want (35.6, 139.7)", latitude, longitude)
		}
		if units != "metric" {
			t.Errorf("FetchWeather() units = %q, want %q", units, "metric")
		}
		if forecast {
			t.Error("FetchWeather() forecast = true, want false")
		}
		return weather.Weather{Current: sampleCurrentWeather()}, nil
	}

	var out bytes.Buffer
	if err := run(&out, "Tokyo", false, "metric", true); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if got := out.String(); !strings.Contains(got, "Location: Tokyo, Japan") {
		t.Errorf("run() output = %q, want it to contain the location", got)
	}
}

func TestRunDetectsLocationWhenNoCityGiven(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(string) (weather.Location, error) {
		t.Error("GeocodeCity() should not be called when no city is given")
		return weather.Location{}, nil
	}
	locateByIP = func() (weather.Location, error) {
		return weather.Location{City: "Kathmandu", Country: "Nepal", Latitude: 27.7, Longitude: 85.3}, nil
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		return weather.Weather{Current: sampleCurrentWeather()}, nil
	}

	var out bytes.Buffer
	if err := run(&out, "", false, "metric", true); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if got := out.String(); !strings.Contains(got, "Location: Kathmandu, Nepal") {
		t.Errorf("run() output = %q, want it to contain the detected location", got)
	}
}

func TestRunRendersForecastWhenDailyIsAvailable(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(string) (weather.Location, error) {
		return weather.Location{City: "Tokyo", Country: "Japan"}, nil
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		return weather.Weather{
			Current: sampleCurrentWeather(),
			Daily:   []weather.DailyForecast{{Date: "2026-08-01", Code: 3, TempMax: 30, TempMin: 20}},
		}, nil
	}

	var out bytes.Buffer
	if err := run(&out, "Tokyo", true, "metric", true); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if got := out.String(); !strings.Contains(got, "5-Day Forecast:") {
		t.Errorf("run() output = %q, want it to contain the forecast", got)
	}
}

func TestRunSkipsForecastWhenDailyIsNil(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(string) (weather.Location, error) {
		return weather.Location{City: "Tokyo", Country: "Japan"}, nil
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		return weather.Weather{Current: sampleCurrentWeather()}, nil
	}

	var out bytes.Buffer
	if err := run(&out, "Tokyo", true, "metric", true); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if got := out.String(); strings.Contains(got, "5-Day Forecast") {
		t.Errorf("run() output = %q, want it to skip an empty forecast", got)
	}
}

func TestRunReturnsErrorWhenGeocodingFails(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(string) (weather.Location, error) {
		return weather.Location{}, weather.ErrCityNotFound
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		t.Error("FetchWeather() should not be called when geocoding fails")
		return weather.Weather{}, nil
	}

	var out bytes.Buffer
	err := run(&out, "Nowhereville", false, "metric", true)
	if !errors.Is(err, weather.ErrCityNotFound) {
		t.Errorf("run() error = %v, want %v", err, weather.ErrCityNotFound)
	}
}

func TestRunReturnsErrorWhenWeatherFetchFails(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(string) (weather.Location, error) {
		return weather.Location{City: "Tokyo", Country: "Japan"}, nil
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		return weather.Weather{}, weather.ErrWeatherService
	}

	var out bytes.Buffer
	err := run(&out, "Tokyo", false, "metric", true)
	if !errors.Is(err, weather.ErrWeatherService) {
		t.Errorf("run() error = %v, want %v", err, weather.ErrWeatherService)
	}
}

func TestNewRootCmdRunsWithValidUnits(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	geocodeCity = func(string) (weather.Location, error) {
		return weather.Location{City: "Tokyo", Country: "Japan"}, nil
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		return weather.Weather{Current: sampleCurrentWeather()}, nil
	}

	cmd := newRootCmd("test")

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--plain", "Tokyo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Location: Tokyo, Japan") {
		t.Errorf("output = %q, want it to contain the rendered location", got)
	}
}

func TestNewRootCmdDetectsLocationWithoutCity(t *testing.T) {
	stubSpinner(t)
	preserveWeatherSeams(t)

	locateByIP = func() (weather.Location, error) {
		return weather.Location{City: "Kathmandu", Country: "Nepal"}, nil
	}
	fetchWeather = func(float64, float64, string, bool) (weather.Weather, error) {
		return weather.Weather{Current: sampleCurrentWeather()}, nil
	}

	cmd := newRootCmd("test")

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--plain"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Location: Kathmandu, Nepal") {
		t.Errorf("output = %q, want it to contain the detected location", got)
	}
}

func sampleCurrentWeather() weather.CurrentWeather {
	return weather.CurrentWeather{
		Temperature: 21,
		FeelsLike:   23,
		Humidity:    60,
		Wind:        5,
		Code:        3,
		CloudCover:  40,
		Time:        "2026-08-01T14:30",
		Sunrise:     "2026-08-01T05:45",
		Sunset:      "2026-08-01T19:10",
	}
}
