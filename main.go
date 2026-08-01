package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ernilambar/skycast/internal/render"
	"github.com/ernilambar/skycast/internal/spinner"
	"github.com/ernilambar/skycast/internal/weather"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	var forecast bool
	var units string
	var plain bool

	rootCmd := &cobra.Command{
		Use:           "skycast [city]",
		Short:         "A terminal weather app",
		Version:       version,
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if units != "metric" && units != "imperial" {
				err := fmt.Errorf(`invalid units "%s": must be "metric" or "imperial"`, units)
				printError(err)
				return err
			}

			var city string
			if len(args) > 0 {
				city = args[0]
			}

			if err := run(city, forecast, units, plain); err != nil {
				printError(err)
				return err
			}

			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&forecast, "forecast", "f", false, "Show 5-day forecast")
	rootCmd.Flags().StringVarP(&units, "units", "u", "metric", "Temperature units: metric (C) or imperial (F)")
	rootCmd.Flags().BoolVarP(&plain, "plain", "p", false, "Output simple plain text without ASCII borders")
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(city string, forecast bool, units string, plain bool) error {
	sp := spinner.New()

	if city != "" {
		sp.Start(fmt.Sprintf("Looking up %s...", city))
	} else {
		sp.Start("Detecting your location...")
	}

	var location weather.Location
	var err error

	if city != "" {
		location, err = weather.GeocodeCity(city)
	} else {
		location, err = weather.LocateByIP()
	}

	if err != nil {
		sp.Stop()
		return err
	}

	sp.SetText("Fetching weather data...")

	w, err := weather.FetchWeather(location.Latitude, location.Longitude, units, forecast)

	sp.Stop()

	if err != nil {
		return err
	}

	render.CurrentWeather(location, w.Current, units, plain)

	if forecast && w.Daily != nil {
		render.Forecast(w.Daily, units, plain)
	}

	return nil
}

func printError(err error) {
	fmt.Fprintln(os.Stderr, render.Red("Error: "+err.Error()))
}
