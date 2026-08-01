package weather

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type stubTransport struct {
	body       string
	statusCode int
}

func (s stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	status := s.statusCode
	if status == 0 {
		status = 200
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

func withMockResponse(t *testing.T, body string, statusCode int) {
	t.Helper()

	original := httpClient
	httpClient = &http.Client{Transport: stubTransport{body: body, statusCode: statusCode}}
	t.Cleanup(func() { httpClient = original })
}

func TestGeocodeCityReturnsFirstMatchingResult(t *testing.T) {
	withMockResponse(t, `{
		"results": [
			{"name": "Tokyo", "country": "Japan", "latitude": 35.6895, "longitude": 139.6917}
		]
	}`, 200)

	location, err := GeocodeCity("Tokyo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Location{City: "Tokyo", Country: "Japan", Latitude: 35.6895, Longitude: 139.6917}
	if !reflect.DeepEqual(location, want) {
		t.Errorf("GeocodeCity() = %+v, want %+v", location, want)
	}
}

func TestGeocodeCityThrowsWhenNoResultsAreFound(t *testing.T) {
	withMockResponse(t, `{"results": []}`, 200)

	_, err := GeocodeCity("Nowhereville")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("GeocodeCity() error = %v, want it to mention \"not found\"", err)
	}
}

func TestLocateByIPReturnsLocationFromIPLookup(t *testing.T) {
	withMockResponse(t, `{
		"city": "Kathmandu", "country_name": "Nepal", "latitude": 27.7172, "longitude": 85.324
	}`, 200)

	location, err := LocateByIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Location{City: "Kathmandu", Country: "Nepal", Latitude: 27.7172, Longitude: 85.324}
	if !reflect.DeepEqual(location, want) {
		t.Errorf("LocateByIP() = %+v, want %+v", location, want)
	}
}

func TestLocateByIPThrowsWhenTheIPCannotBeResolved(t *testing.T) {
	withMockResponse(t, `{"error": true, "reason": "RateLimited"}`, 200)

	_, err := LocateByIP()
	if err == nil || !strings.Contains(err.Error(), "unable to detect location") {
		t.Errorf("LocateByIP() error = %v, want it to mention \"unable to detect location\"", err)
	}
}

func TestFetchWeatherParsesCurrentWeather(t *testing.T) {
	withMockResponse(t, `{
		"current": {
			"temperature_2m": 21,
			"relative_humidity_2m": 60,
			"apparent_temperature": 23,
			"weather_code": 3,
			"wind_speed_10m": 5
		}
	}`, 200)

	w, err := FetchWeather(27.7, 85.3, "metric", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Weather{Current: CurrentWeather{Temperature: 21, FeelsLike: 23, Humidity: 60, Wind: 5, Code: 3}}
	if !reflect.DeepEqual(w, want) {
		t.Errorf("FetchWeather() = %+v, want %+v", w, want)
	}
}

func TestFetchWeatherParsesTheDailyForecastWhenRequested(t *testing.T) {
	withMockResponse(t, `{
		"current": {
			"temperature_2m": 21,
			"relative_humidity_2m": 60,
			"apparent_temperature": 23,
			"weather_code": 3,
			"wind_speed_10m": 5
		},
		"daily": {
			"time": ["2026-08-01", "2026-08-02"],
			"weather_code": [3, 61],
			"temperature_2m_max": [30, 28],
			"temperature_2m_min": [20, 19]
		}
	}`, 200)

	w, err := FetchWeather(27.7, 85.3, "metric", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []DailyForecast{
		{Date: "2026-08-01", Code: 3, TempMax: 30, TempMin: 20},
		{Date: "2026-08-02", Code: 61, TempMax: 28, TempMin: 19},
	}
	if !reflect.DeepEqual(w.Daily, want) {
		t.Errorf("FetchWeather().Daily = %+v, want %+v", w.Daily, want)
	}
}

func TestFetchWeatherThrowsWhenTheAPIRespondsWithAnErrorStatus(t *testing.T) {
	withMockResponse(t, `{}`, 500)

	_, err := FetchWeather(0, 0, "metric", false)
	if err == nil || !strings.Contains(err.Error(), "weather service") {
		t.Errorf("FetchWeather() error = %v, want it to mention \"weather service\"", err)
	}
}
