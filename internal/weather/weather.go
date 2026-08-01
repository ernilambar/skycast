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
}

// DailyForecast holds one day of a multi-day forecast.
type DailyForecast struct {
	Date    string
	Code    int
	TempMax float64
	TempMin float64
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

// LocateByIP resolves the caller's location from their IP address.
func LocateByIP() (Location, error) {
	res, err := httpClient.Get("https://ipapi.co/json/")
	if err != nil {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}

	var data struct {
		City        string  `json:"city"`
		CountryName string  `json:"country_name"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Error       bool    `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Location{}, fmt.Errorf("unable to reach the IP location service")
	}

	if data.Error || data.Latitude == 0 {
		return Location{}, fmt.Errorf("unable to detect location from IP address")
	}

	return Location{
		City:      data.City,
		Country:   data.CountryName,
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
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
		"current":          {"temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m"},
		"temperature_unit": {temperatureUnit},
		"wind_speed_unit":  {windSpeedUnit},
		"timezone":         {"auto"},
	}

	if forecast {
		params.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min")
		params.Set("forecast_days", "5")
	}

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
			Temperature2m       float64 `json:"temperature_2m"`
			RelativeHumidity2m  float64 `json:"relative_humidity_2m"`
			ApparentTemperature float64 `json:"apparent_temperature"`
			WeatherCode         int     `json:"weather_code"`
			WindSpeed10m        float64 `json:"wind_speed_10m"`
		} `json:"current"`
		Daily struct {
			Time             []string  `json:"time"`
			WeatherCode      []int     `json:"weather_code"`
			Temperature2mMax []float64 `json:"temperature_2m_max"`
			Temperature2mMin []float64 `json:"temperature_2m_min"`
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
		},
	}

	if forecast && len(data.Daily.Time) > 0 {
		result.Daily = make([]DailyForecast, len(data.Daily.Time))
		for i, date := range data.Daily.Time {
			result.Daily[i] = DailyForecast{
				Date:    date,
				Code:    data.Daily.WeatherCode[i],
				TempMax: data.Daily.Temperature2mMax[i],
				TempMin: data.Daily.Temperature2mMin[i],
			}
		}
	}

	return result, nil
}
