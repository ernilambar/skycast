package weather

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// stubTransport is a mock http.RoundTripper that records every request it
// receives so tests can assert on the outgoing method, host and query params.
type stubTransport struct {
	mu       sync.Mutex
	requests []*http.Request
	body     string
	status   int
	err      error
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	s.requests = append(s.requests, req)
	s.mu.Unlock()

	if s.err != nil {
		return nil, s.err
	}

	status := s.status
	if status == 0 {
		status = http.StatusOK
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

// lastRequest returns the most recent request the stub handled, failing the
// test if no request was ever issued.
func (s *stubTransport) lastRequest(t *testing.T) *http.Request {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.requests) == 0 {
		t.Fatal("expected a request to be issued, but none was")
	}

	return s.requests[len(s.requests)-1]
}

func withMockResponse(t *testing.T, body string, statusCode int) *stubTransport {
	t.Helper()

	stub := &stubTransport{body: body, status: statusCode}
	original := httpClient
	httpClient = &http.Client{Transport: stub}
	t.Cleanup(func() { httpClient = original })

	return stub
}

func withMockError(t *testing.T, err error) *stubTransport {
	t.Helper()

	stub := &stubTransport{err: err}
	original := httpClient
	httpClient = &http.Client{Transport: stub}
	t.Cleanup(func() { httpClient = original })

	return stub
}

// hostStubTransport routes each request to a per-host stub so tests can make
// one provider fail while another succeeds.
type hostStubTransport struct {
	mu     sync.Mutex
	byHost map[string]*stubTransport
}

func (h *hostStubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	h.mu.Lock()
	stub, ok := h.byHost[req.URL.Host]
	h.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("no stub registered for host %q", req.URL.Host)
	}

	return stub.RoundTrip(req)
}

func withMockResponsesByHost(t *testing.T, byHost map[string]*stubTransport) {
	t.Helper()

	original := httpClient
	httpClient = &http.Client{Transport: &hostStubTransport{byHost: byHost}}
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

func TestGeocodeCitySendsNameAndCountQueryParams(t *testing.T) {
	stub := withMockResponse(t, `{"results": [{"name": "Tokyo"}]}`, 200)

	if _, err := GeocodeCity("Tokyo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := stub.lastRequest(t)

	if req.Method != http.MethodGet {
		t.Errorf("request method = %q, want %q", req.Method, http.MethodGet)
	}
	if req.URL.Host != "geocoding-api.open-meteo.com" {
		t.Errorf("request host = %q, want %q", req.URL.Host, "geocoding-api.open-meteo.com")
	}

	query := req.URL.Query()
	if got := query.Get("name"); got != "Tokyo" {
		t.Errorf("name param = %q, want %q", got, "Tokyo")
	}
	if got := query.Get("count"); got != "1" {
		t.Errorf("count param = %q, want %q", got, "1")
	}
}

func TestLocateByIPReturnsLocationFromIPLookup(t *testing.T) {
	withMockResponse(t, `{
		"city": "Kathmandu", "country": "Nepal", "latitude": 27.7172, "longitude": 85.324, "success": true
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

func TestLocateByIPFallsBackToSecondProviderWhenPrimaryFails(t *testing.T) {
	withMockResponsesByHost(t, map[string]*stubTransport{
		"ipwho.is":   {status: 429},
		"ip-api.com": {body: `{"status": "success", "city": "Kathmandu", "country": "Nepal", "lat": 27.7172, "lon": 85.324}`},
	})

	location, err := LocateByIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Location{City: "Kathmandu", Country: "Nepal", Latitude: 27.7172, Longitude: 85.324}
	if !reflect.DeepEqual(location, want) {
		t.Errorf("LocateByIP() = %+v, want %+v", location, want)
	}
}

func TestLocateByIPFallsBackWhenPrimaryReturnsZeroCoordinates(t *testing.T) {
	withMockResponsesByHost(t, map[string]*stubTransport{
		"ipwho.is":   {body: `{"success": true, "city": "Nowhere", "latitude": 0, "longitude": 0}`},
		"ip-api.com": {body: `{"status": "success", "city": "Kathmandu", "country": "Nepal", "lat": 27.7172, "lon": 85.324}`},
	})

	location, err := LocateByIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Location{City: "Kathmandu", Country: "Nepal", Latitude: 27.7172, Longitude: 85.324}
	if !reflect.DeepEqual(location, want) {
		t.Errorf("LocateByIP() = %+v, want %+v", location, want)
	}
}

func TestFetchWeatherParsesCurrentWeather(t *testing.T) {
	withMockResponse(t, `{
		"current": {
			"time": "2026-08-01T14:30",
			"temperature_2m": 21,
			"relative_humidity_2m": 60,
			"apparent_temperature": 23,
			"weather_code": 3,
			"wind_speed_10m": 5,
			"cloud_cover": 40,
			"precipitation": 1.2
		},
		"daily": {
			"sunrise": ["2026-08-01T05:45"],
			"sunset": ["2026-08-01T19:10"]
		}
	}`, 200)

	w, err := FetchWeather(27.7, 85.3, "metric", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Weather{Current: CurrentWeather{
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
	}}
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
			"temperature_2m_min": [20, 19],
			"precipitation_probability_max": [10, 55]
		}
	}`, 200)

	w, err := FetchWeather(27.7, 85.3, "metric", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []DailyForecast{
		{Date: "2026-08-01", Code: 3, TempMax: 30, TempMin: 20, PrecipitationProbability: 10},
		{Date: "2026-08-02", Code: 61, TempMax: 28, TempMin: 19, PrecipitationProbability: 55},
	}
	if !reflect.DeepEqual(w.Daily, want) {
		t.Errorf("FetchWeather().Daily = %+v, want %+v", w.Daily, want)
	}
}

func TestFetchWeatherToleratesForecastWithoutPrecipitationProbability(t *testing.T) {
	withMockResponse(t, `{
		"current": {"weather_code": 3},
		"daily": {
			"time": ["2026-08-01"],
			"weather_code": [3],
			"temperature_2m_max": [30],
			"temperature_2m_min": [20]
		}
	}`, 200)

	w, err := FetchWeather(27.7, 85.3, "metric", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.Daily) != 1 {
		t.Fatalf("FetchWeather().Daily has %d entries, want 1", len(w.Daily))
	}
	if w.Daily[0].PrecipitationProbability != 0 {
		t.Errorf("PrecipitationProbability = %v, want 0", w.Daily[0].PrecipitationProbability)
	}
}

func TestFetchWeatherSkipsForecastWhenNotRequested(t *testing.T) {
	withMockResponse(t, `{
		"current": {"weather_code": 3},
		"daily": {
			"time": ["2026-08-01"],
			"weather_code": [3],
			"temperature_2m_max": [30],
			"temperature_2m_min": [20]
		}
	}`, 200)

	w, err := FetchWeather(27.7, 85.3, "metric", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Daily != nil {
		t.Errorf("FetchWeather().Daily = %+v, want nil when forecast is not requested", w.Daily)
	}
}

func TestFetchWeatherSendsMetricParamsByDefault(t *testing.T) {
	stub := withMockResponse(t, `{"current": {"weather_code": 3}}`, 200)

	if _, err := FetchWeather(27.7, 85.3, "metric", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := stub.lastRequest(t).URL.Query()

	if got := query.Get("latitude"); got != "27.7" {
		t.Errorf("latitude param = %q, want %q", got, "27.7")
	}
	if got := query.Get("longitude"); got != "85.3" {
		t.Errorf("longitude param = %q, want %q", got, "85.3")
	}
	if got := query.Get("temperature_unit"); got != "celsius" {
		t.Errorf("temperature_unit param = %q, want %q", got, "celsius")
	}
	if got := query.Get("wind_speed_unit"); got != "kmh" {
		t.Errorf("wind_speed_unit param = %q, want %q", got, "kmh")
	}
	if got := query.Get("precipitation_unit"); got != "" {
		t.Errorf("precipitation_unit param = %q, want it to be unset for metric", got)
	}
	if got := query.Get("forecast_days"); got != "1" {
		t.Errorf("forecast_days param = %q, want %q", got, "1")
	}
	if got := query.Get("daily"); got != "sunrise,sunset" {
		t.Errorf("daily param = %q, want %q", got, "sunrise,sunset")
	}
}

func TestFetchWeatherSendsImperialParams(t *testing.T) {
	stub := withMockResponse(t, `{"current": {"weather_code": 3}}`, 200)

	if _, err := FetchWeather(27.7, 85.3, "imperial", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := stub.lastRequest(t).URL.Query()

	if got := query.Get("temperature_unit"); got != "fahrenheit" {
		t.Errorf("temperature_unit param = %q, want %q", got, "fahrenheit")
	}
	if got := query.Get("wind_speed_unit"); got != "mph" {
		t.Errorf("wind_speed_unit param = %q, want %q", got, "mph")
	}
	if got := query.Get("precipitation_unit"); got != "inch" {
		t.Errorf("precipitation_unit param = %q, want %q", got, "inch")
	}
}

func TestFetchWeatherRequestsFiveDaysAndExtendedDailyFieldsForForecast(t *testing.T) {
	stub := withMockResponse(t, `{"current": {"weather_code": 3}}`, 200)

	if _, err := FetchWeather(27.7, 85.3, "metric", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := stub.lastRequest(t).URL.Query()

	if got := query.Get("forecast_days"); got != "5" {
		t.Errorf("forecast_days param = %q, want %q", got, "5")
	}

	daily := query.Get("daily")
	for _, field := range []string{"weather_code", "temperature_2m_max", "temperature_2m_min", "precipitation_probability_max"} {
		if !strings.Contains(daily, field) {
			t.Errorf("daily param = %q, want it to contain %q", daily, field)
		}
	}
}

func TestGeocodeCityReturnsErrorWhenServiceUnreachable(t *testing.T) {
	withMockError(t, errors.New("connection refused"))

	_, err := GeocodeCity("Tokyo")
	if !errors.Is(err, ErrGeocodingService) {
		t.Errorf("GeocodeCity() error = %v, want %v", err, ErrGeocodingService)
	}
}

func TestGeocodeCityReturnsErrorOnNon2xxStatus(t *testing.T) {
	withMockResponse(t, `{}`, 500)

	_, err := GeocodeCity("Tokyo")
	if !errors.Is(err, ErrGeocodingService) {
		t.Errorf("GeocodeCity() error = %v, want %v", err, ErrGeocodingService)
	}
}

func TestGeocodeCityReturnsErrorOnMalformedJSON(t *testing.T) {
	withMockResponse(t, `{not json`, 200)

	_, err := GeocodeCity("Tokyo")
	if !errors.Is(err, ErrGeocodingService) {
		t.Errorf("GeocodeCity() error = %v, want %v", err, ErrGeocodingService)
	}
}

func TestGeocodeCityReturnsErrorWhenNoResultsFound(t *testing.T) {
	withMockResponse(t, `{"results": []}`, 200)

	_, err := GeocodeCity("Nowhereville")
	if !errors.Is(err, ErrCityNotFound) {
		t.Errorf("GeocodeCity() error = %v, want %v", err, ErrCityNotFound)
	}
	if !strings.Contains(err.Error(), "Nowhereville") {
		t.Errorf("GeocodeCity() error = %v, want it to mention the city name", err)
	}
}

func TestGeocodeCityReturnsErrorWhenResultsAreNull(t *testing.T) {
	withMockResponse(t, `{"results": null}`, 200)

	_, err := GeocodeCity("Nowhereville")
	if !errors.Is(err, ErrCityNotFound) {
		t.Errorf("GeocodeCity() error = %v, want %v", err, ErrCityNotFound)
	}
}

func TestLocateByIPReturnsErrorWhenLocationCannotBeDetected(t *testing.T) {
	withMockResponse(t, `{"success": false}`, 200)

	_, err := LocateByIP()
	if !errors.Is(err, ErrIPDetection) {
		t.Errorf("LocateByIP() error = %v, want %v", err, ErrIPDetection)
	}
}

func TestLocateByIPReturnsErrorWhenBothProvidersFail(t *testing.T) {
	withMockResponsesByHost(t, map[string]*stubTransport{
		"ipwho.is":   {status: 429},
		"ip-api.com": {status: 429},
	})

	_, err := LocateByIP()
	if !errors.Is(err, ErrIPLocationService) {
		t.Errorf("LocateByIP() error = %v, want %v", err, ErrIPLocationService)
	}
}

func TestLocateByIPReturnsErrorWhenServiceUnreachable(t *testing.T) {
	withMockError(t, errors.New("dial tcp: no route to host"))

	_, err := LocateByIP()
	if !errors.Is(err, ErrIPLocationService) {
		t.Errorf("LocateByIP() error = %v, want %v", err, ErrIPLocationService)
	}
}

func TestLocateByIPFallsBackWhenPrimaryReturnsMalformedJSON(t *testing.T) {
	withMockResponsesByHost(t, map[string]*stubTransport{
		"ipwho.is":   {body: `{not json`},
		"ip-api.com": {body: `{"status": "success", "city": "Kathmandu", "country": "Nepal", "lat": 27.7172, "lon": 85.324}`},
	})

	location, err := LocateByIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Location{City: "Kathmandu", Country: "Nepal", Latitude: 27.7172, Longitude: 85.324}
	if !reflect.DeepEqual(location, want) {
		t.Errorf("LocateByIP() = %+v, want %+v", location, want)
	}
}

func TestLocateByIPReturnsErrorWhenBothProvidersReturnMalformedJSON(t *testing.T) {
	withMockResponsesByHost(t, map[string]*stubTransport{
		"ipwho.is":   {body: `{not json`},
		"ip-api.com": {body: `{not json`},
	})

	_, err := LocateByIP()
	if !errors.Is(err, ErrIPLocationService) {
		t.Errorf("LocateByIP() error = %v, want %v", err, ErrIPLocationService)
	}
}

func TestFetchWeatherReturnsErrorWhenServiceUnreachable(t *testing.T) {
	withMockError(t, errors.New("connection reset"))

	_, err := FetchWeather(0, 0, "metric", false)
	if !errors.Is(err, ErrWeatherService) {
		t.Errorf("FetchWeather() error = %v, want %v", err, ErrWeatherService)
	}
}

func TestFetchWeatherReturnsErrorOnErrorStatus(t *testing.T) {
	withMockResponse(t, `{}`, 500)

	_, err := FetchWeather(0, 0, "metric", false)
	if !errors.Is(err, ErrWeatherService) {
		t.Errorf("FetchWeather() error = %v, want %v", err, ErrWeatherService)
	}
}

func TestFetchWeatherReturnsErrorOnMalformedJSON(t *testing.T) {
	withMockResponse(t, `{not json`, 200)

	_, err := FetchWeather(0, 0, "metric", false)
	if !errors.Is(err, ErrWeatherService) {
		t.Errorf("FetchWeather() error = %v, want %v", err, ErrWeatherService)
	}
}
