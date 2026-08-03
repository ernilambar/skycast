package render

var weatherIcons = map[string]string{
	"Clear": Yellow(
		"\n    \\   /\n     .-.\n  ― (   ) ―\n     `-'\n    /   \\     ",
	),
	"Rain": Cyan(
		"\n     .--.\n    (    ).\n   (___.__)\n    ʻ ʻ ʻ ʻ\n   ʻ ʻ ʻ ʻ    ",
	),
	"Clouds": Gray(
		"\n      .--.\n   .-(    ).\n  (___.__)__)\n              ",
	),
	"Snow": White(
		"\n     .--.\n    (    ).\n   (___.__)\n    *  *  *\n   *  *  *    ",
	),
	"Thunderstorm": Yellow(
		"\n     .--.\n    (    ).\n   (___.__)\n   ⚡  ʻ  ⚡\n    ʻ ʻ ʻ ʻ    ",
	),
	"Fog": Gray(
		"\n   _ - _ - _\n  _ - _ - _ -\n   _ - _ - _\n              ",
	),
}

var conditionEmoji = map[string]string{
	"Clear":        "☀️",
	"Clouds":       "☁️",
	"Rain":         "🌧️",
	"Snow":         "❄️",
	"Thunderstorm": "⛈️",
	"Fog":          "🌫️",
}

var timeOfDayEmoji = map[string]string{
	"Dawn":  "🌅",
	"Day":   "☀️",
	"Dusk":  "🌇",
	"Night": "🌙",
}

var snowCodes = map[int]bool{71: true, 73: true, 75: true, 77: true, 85: true, 86: true}

var rainCodes = map[int]bool{
	51: true, 53: true, 55: true, 56: true, 57: true,
	61: true, 63: true, 65: true, 66: true, 67: true,
	80: true, 81: true, 82: true,
}

// CodeToCondition maps a WMO weather code to a broad condition name.
func CodeToCondition(code int) string {
	switch {
	case code == 0 || code == 1:
		return "Clear"
	case code == 2 || code == 3:
		return "Clouds"
	case code == 45 || code == 48:
		return "Fog"
	case code >= 95:
		return "Thunderstorm"
	case snowCodes[code]:
		return "Snow"
	case rainCodes[code]:
		return "Rain"
	default:
		return "Clouds"
	}
}
