package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var httpClient = &http.Client{}

// Location is a resolved place with coordinates.
type Location struct {
	City      string
	Country   string
	Latitude  float64
	Longitude float64
}

// CurrentWeather holds the current conditions for a location.
type CurrentWeather struct {
	Temperature float64
	FeelsLike   float64
	Humidity    float64
	Wind        float64
	Code        int
	CloudCover  float64
	Time        string
	Sunrise     string
	Sunset      string
}

// DailyForecast holds one day of a multi-day forecast.
type DailyForecast struct {
	Date                     string
	Code                     int
	TempMax                  float64
	TempMin                  float64
	PrecipitationProbability float64
}

// Weather bundles current conditions with an optional daily forecast.
type Weather struct {
	Current CurrentWeather
	Daily   []DailyForecast
}

// GeocodeCity resolves a city name to coordinates.
func GeocodeCity(city string) (Location, error) {
	u := "https://geocoding-api.open-meteo.com/v1/search?" + url.Values{
		"name":  {city},
		"count": {"1"},
	}.Encode()

	res, err := httpClient.Get(u)
	if err != nil {
		return Location{}, fmt.Errorf("unable to reach the geocoding service")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Location{}, fmt.Errorf("unable to reach the geocoding service")
	}

	var data struct {
		Results []struct {
			Name      string  `json:"name"`
			Country   string  `json:"country"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Location{}, fmt.Errorf("unable to reach the geocoding service")
	}

	if len(data.Results) == 0 {
		return Location{}, fmt.Errorf("city %q not found", city)
	}

	result := data.Results[0]

	return Location{
		City:      result.Name,
		Country:   result.Country,
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
	}, nil
}

// LocateByIP resolves the caller's location from their IP address, trying
// each provider in turn until one succeeds.
func LocateByIP() (Location, error) {
	providers := []func() (Location, error){locateByIPWhoIs, locateByIPAPI}

	var err error
	for _, provider := range providers {
		var location Location

		location, err = provider()
		if err == nil {
			return location, nil
		}
	}

	return Location{}, err
}

func locateByIPWhoIs() (Location, error) {
	res, err := httpClient.Get("https://ipwho.is/")
	if err != nil {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}

	var data struct {
		City      string  `json:"city"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Success   bool    `json:"success"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}

	if !data.Success || data.Latitude == 0 {
		return Location{}, fmt.Errorf("unable to detect location from IP address")
	}

	return Location{
		City:      data.City,
		Country:   data.Country,
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
	}, nil
}

func locateByIPAPI() (Location, error) {
	res, err := httpClient.Get("http://ip-api.com/json/")
	if err != nil {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}

	var data struct {
		Status  string  `json:"status"`
		City    string  `json:"city"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}

	if data.Status != "success" || data.Lat == 0 {
		return Location{}, fmt.Errorf("unable to detect location from IP address")
	}

	return Location{
		City:      data.City,
		Country:   data.Country,
		Latitude:  data.Lat,
		Longitude: data.Lon,
	}, nil
}

// FetchWeather fetches current conditions, and optionally a 5-day forecast, for a location.
func FetchWeather(latitude, longitude float64, units string, forecast bool) (Weather, error) {
	temperatureUnit := "celsius"
	windSpeedUnit := "kmh"

	if units == "imperial" {
		temperatureUnit = "fahrenheit"
		windSpeedUnit = "mph"
	}

	params := url.Values{
		"latitude":         {fmt.Sprintf("%g", latitude)},
		"longitude":        {fmt.Sprintf("%g", longitude)},
		"current":          {"temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m,cloud_cover"},
		"temperature_unit": {temperatureUnit},
		"wind_speed_unit":  {windSpeedUnit},
		"timezone":         {"auto"},
	}

	dailyFields := "sunrise,sunset"
	forecastDays := "1"

	if forecast {
		dailyFields = "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,sunrise,sunset"
		forecastDays = "5"
	}

	params.Set("daily", dailyFields)
	params.Set("forecast_days", forecastDays)

	res, err := httpClient.Get("https://api.open-meteo.com/v1/forecast?" + params.Encode())
	if err != nil {
		return Weather{}, fmt.Errorf("unable to reach the weather service")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Weather{}, fmt.Errorf("unable to reach the weather service")
	}

	var data struct {
		Current struct {
			Time                string  `json:"time"`
			Temperature2m       float64 `json:"temperature_2m"`
			RelativeHumidity2m  float64 `json:"relative_humidity_2m"`
			ApparentTemperature float64 `json:"apparent_temperature"`
			WeatherCode         int     `json:"weather_code"`
			WindSpeed10m        float64 `json:"wind_speed_10m"`
			CloudCover          float64 `json:"cloud_cover"`
		} `json:"current"`
		Daily struct {
			Time                        []string  `json:"time"`
			WeatherCode                 []int     `json:"weather_code"`
			Temperature2mMax            []float64 `json:"temperature_2m_max"`
			Temperature2mMin            []float64 `json:"temperature_2m_min"`
			PrecipitationProbabilityMax []float64 `json:"precipitation_probability_max"`
			Sunrise                     []string  `json:"sunrise"`
			Sunset                      []string  `json:"sunset"`
		} `json:"daily"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Weather{}, fmt.Errorf("unable to reach the weather service")
	}

	result := Weather{
		Current: CurrentWeather{
			Temperature: data.Current.Temperature2m,
			FeelsLike:   data.Current.ApparentTemperature,
			Humidity:    data.Current.RelativeHumidity2m,
			Wind:        data.Current.WindSpeed10m,
			Code:        data.Current.WeatherCode,
			CloudCover:  data.Current.CloudCover,
			Time:        data.Current.Time,
		},
	}

	if len(data.Daily.Sunrise) > 0 {
		result.Current.Sunrise = data.Daily.Sunrise[0]
	}
	if len(data.Daily.Sunset) > 0 {
		result.Current.Sunset = data.Daily.Sunset[0]
	}

	if forecast && len(data.Daily.Time) > 0 {
		result.Daily = make([]DailyForecast, len(data.Daily.Time))
		for i, date := range data.Daily.Time {
			var precipitation float64
			if i < len(data.Daily.PrecipitationProbabilityMax) {
				precipitation = data.Daily.PrecipitationProbabilityMax[i]
			}

			result.Daily[i] = DailyForecast{
				Date:                     date,
				Code:                     data.Daily.WeatherCode[i],
				TempMax:                  data.Daily.Temperature2mMax[i],
				TempMin:                  data.Daily.Temperature2mMin[i],
				PrecipitationProbability: precipitation,
			}
		}
	}

	return result, nil
}
