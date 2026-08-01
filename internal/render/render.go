package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ernilambar/skycast/internal/weather"
)

func unitLabel(units string) string {
	if units == "imperial" {
		return "°F"
	}
	return "°C"
}

func windUnitLabel(units string) string {
	if units == "imperial" {
		return "mph"
	}
	return "km/h"
}

func round(v float64) int {
	return int(math.Round(v))
}

// numStr formats a float the way JavaScript's default number-to-string
// conversion would, so values like humidity/wind print without a
// spurious ".0" for whole numbers.
func numStr(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatTemp(value float64, units string) string {
	text := fmt.Sprintf("%d%s", round(value), unitLabel(units))

	celsius := value
	if units == "imperial" {
		celsius = (value - 32) * 5 / 9
	}

	switch {
	case celsius <= 0:
		return BoldCyan(text + " ❄️")
	case celsius <= 18:
		return BoldBlue(text + " 🌤️")
	case celsius <= 28:
		return BoldYellow(text + " ☀️")
	default:
		return BoldRed(text + " 🔥")
	}
}

func locationLabel(location weather.Location) string {
	if location.Country != "" {
		return location.City + ", " + location.Country
	}
	return location.City
}

func weekdayName(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Mon")
}

// CurrentWeather prints the current conditions for a location, either as a
// bordered card or as plain text.
func CurrentWeather(location weather.Location, current weather.CurrentWeather, units string, plain bool) {
	condition := CodeToCondition(current.Code)

	if plain {
		fmt.Printf("Location: %s\n", locationLabel(location))
		fmt.Printf("Condition: %s\n", condition)
		fmt.Printf("Temperature: %d%s\n", round(current.Temperature), unitLabel(units))
		fmt.Printf("Feels Like: %d%s\n", round(current.FeelsLike), unitLabel(units))
		fmt.Printf("Humidity: %s%%\n", numStr(current.Humidity))
		fmt.Printf("Wind Speed: %s %s\n", numStr(current.Wind), windUnitLabel(units))
		return
	}

	icon := weatherIcons[condition]
	cityTitle := BoldCyan(fmt.Sprintf("  📍 %s  ", strings.ToUpper(locationLabel(location))))

	stats := fmt.Sprintf(
		"\n%s   %s\n%s %s\n%s  %d%s\n%s    %s%%\n%s  %s %s\n",
		Bold("Condition:"), Italic(condition),
		Bold("Temperature:"), formatTemp(current.Temperature, units),
		Bold("Feels Like:"), round(current.FeelsLike), unitLabel(units),
		Bold("Humidity:"), numStr(current.Humidity),
		Bold("Wind Speed:"), numStr(current.Wind), windUnitLabel(units),
	)

	cardBody := icon + "\n" + stats

	fmt.Println(box(cardBody, cityTitle, Cyan, 1, margin{Top: 1, Bottom: 1, Left: 1, Right: 1}))
}

// Forecast prints a multi-day forecast, either as a bordered card or as
// plain text.
func Forecast(daily []weather.DailyForecast, units string, plain bool) {
	if plain {
		fmt.Println()
		fmt.Println("5-Day Forecast:")
		for _, day := range daily {
			condition := CodeToCondition(day.Code)
			fmt.Printf(
				"%s: %s, High %d%s, Low %d%s\n",
				weekdayName(day.Date), condition,
				round(day.TempMax), unitLabel(units),
				round(day.TempMin), unitLabel(units),
			)
		}
		return
	}

	rows := make([]string, len(daily))
	for i, day := range daily {
		condition := CodeToCondition(day.Code)
		label := Bold(fmt.Sprintf("%-4s", weekdayName(day.Date)))
		high := Red(fmt.Sprintf("%5s", fmt.Sprintf("%d°", round(day.TempMax))))
		low := Blue(fmt.Sprintf("%5s", fmt.Sprintf("%d°", round(day.TempMin))))
		rows[i] = fmt.Sprintf(
			"%s %s  %s %s / %s",
			label, conditionEmoji[condition], fmt.Sprintf("%-13s", condition), high, low,
		)
	}

	content := strings.Join(rows, "\n")

	fmt.Println(box(content, BoldCyan("5-Day Forecast"), Cyan, 1, margin{Top: 0, Bottom: 1, Left: 1, Right: 1}))
}
