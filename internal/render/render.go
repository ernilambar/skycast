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

func precipitationUnitLabel(units string) string {
	if units == "imperial" {
		return "in"
	}
	return "mm"
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

func formatClock(iso string) string {
	t, err := time.Parse("2006-01-02T15:04", iso)
	if err != nil {
		return "--:--"
	}
	return t.Format("15:04")
}

// timeOfDayBucket classifies the current local time into dawn/day/dusk/night
// relative to the location's actual sunrise/sunset, using a 30-minute twilight
// window either side. Falls back to "Day" if any timestamp fails to parse.
func timeOfDayBucket(current, sunrise, sunset string) string {
	const layout = "2006-01-02T15:04"

	cur, err := time.Parse(layout, current)
	if err != nil {
		return "Day"
	}
	sr, err := time.Parse(layout, sunrise)
	if err != nil {
		return "Day"
	}
	ss, err := time.Parse(layout, sunset)
	if err != nil {
		return "Day"
	}

	const twilight = 30 * time.Minute
	dawnStart, dawnEnd := sr.Add(-twilight), sr.Add(twilight)
	duskStart, duskEnd := ss.Add(-twilight), ss.Add(twilight)

	switch {
	case !cur.Before(dawnStart) && cur.Before(dawnEnd):
		return "Dawn"
	case !cur.Before(dawnEnd) && cur.Before(duskStart):
		return "Day"
	case !cur.Before(duskStart) && cur.Before(duskEnd):
		return "Dusk"
	default:
		return "Night"
	}
}

// CurrentWeather prints the current conditions for a location, either as a
// bordered card or as plain text.
func CurrentWeather(location weather.Location, current weather.CurrentWeather, units string, plain bool) {
	condition := CodeToCondition(current.Code)
	timeOfDay := timeOfDayBucket(current.Time, current.Sunrise, current.Sunset)

	if plain {
		fmt.Printf("Location: %s\n", locationLabel(location))
		fmt.Printf("Condition: %s\n", condition)
		fmt.Printf("Temperature: %d%s\n", round(current.Temperature), unitLabel(units))
		fmt.Printf("Feels Like: %d%s\n", round(current.FeelsLike), unitLabel(units))
		fmt.Printf("Humidity: %s%%\n", numStr(current.Humidity))
		fmt.Printf("Wind Speed: %s %s\n", numStr(current.Wind), windUnitLabel(units))
		fmt.Printf("Precipitation: %s %s\n", numStr(current.Precipitation), precipitationUnitLabel(units))
		fmt.Printf("Cloud Cover: %s%%\n", numStr(current.CloudCover))
		fmt.Printf("Time of Day: %s\n", timeOfDay)
		fmt.Printf("Sunrise: %s\n", formatClock(current.Sunrise))
		fmt.Printf("Sunset: %s\n", formatClock(current.Sunset))
		return
	}

	icon := weatherIcons[condition]
	cityTitle := BoldCyan(fmt.Sprintf("📍 %s", strings.ToUpper(locationLabel(location))))

	stats := fmt.Sprintf(
		"%s   %s\n%s %s\n%s  %d%s\n%s    %s%%\n%s  %s %s\n%s %s %s\n%s %s%%\n%s %s\n%s     %s\n%s      %s",
		Bold("Condition:"), Italic(condition),
		Bold("Temperature:"), formatTemp(current.Temperature, units),
		Bold("Feels Like:"), round(current.FeelsLike), unitLabel(units),
		Bold("Humidity:"), numStr(current.Humidity),
		Bold("Wind Speed:"), numStr(current.Wind), windUnitLabel(units),
		Bold("Precipitation:"), numStr(current.Precipitation), precipitationUnitLabel(units),
		Bold("Cloud Cover:"), numStr(current.CloudCover),
		Bold("Time of Day:"), fmt.Sprintf("%s %s", timeOfDay, timeOfDayEmoji[timeOfDay]),
		Bold("Sunrise:"), formatClock(current.Sunrise),
		Bold("Sunset:"), formatClock(current.Sunset),
	)

	fmt.Println()
	fmt.Println(cityTitle)
	fmt.Println(icon)
	fmt.Println(stats)
	fmt.Println()
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
				"%s: %s, High %d%s, Low %d%s, Precipitation %d%%\n",
				weekdayName(day.Date), condition,
				round(day.TempMax), unitLabel(units),
				round(day.TempMin), unitLabel(units),
				round(day.PrecipitationProbability),
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
		rain := Cyan(fmt.Sprintf("%3d%%", round(day.PrecipitationProbability)))
		rows[i] = fmt.Sprintf(
			"%s %s  %s %s / %s  💧 %s",
			label, conditionEmoji[condition], fmt.Sprintf("%-13s", condition), high, low, rain,
		)
	}

	fmt.Println(BoldCyan("5-Day Forecast"))
	fmt.Println(strings.Join(rows, "\n"))
	fmt.Println()
}
