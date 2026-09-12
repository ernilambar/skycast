package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ernilambar/skycast/internal/render"
	"github.com/ernilambar/skycast/internal/spinner"
	"github.com/ernilambar/skycast/internal/weather"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

// Package-level seams so tests can substitute the network calls and spinner.
// This mirrors the httpClient override pattern used in internal/weather.
var (
	geocodeCity  = weather.GeocodeCity
	locateByIP   = weather.LocateByIP
	fetchWeather = weather.FetchWeather
	newSpinner   = spinner.New
)

func main() {
	if err := newRootCmd(version).Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd(version string) *cobra.Command {
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
				printError(cmd.ErrOrStderr(), err)
				return err
			}

			var city string
			if len(args) > 0 {
				city = args[0]
			}

			if err := run(cmd.OutOrStdout(), city, forecast, units, plain); err != nil {
				printError(cmd.ErrOrStderr(), err)
				return err
			}

			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&forecast, "forecast", "f", false, "Show 5-day forecast")
	rootCmd.Flags().StringVarP(&units, "units", "u", "metric", "Temperature units: metric (C) or imperial (F)")
	rootCmd.Flags().BoolVarP(&plain, "plain", "p", false, "Output simple plain text without colors or icons")
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	return rootCmd
}

func run(out io.Writer, city string, forecast bool, units string, plain bool) error {
	sp := newSpinner()

	if city != "" {
		sp.Start(fmt.Sprintf("Looking up %s...", city))
	} else {
		sp.Start("Detecting your location...")
	}

	var location weather.Location
	var err error

	if city != "" {
		location, err = geocodeCity(city)
	} else {
		location, err = locateByIP()
	}

	if err != nil {
		sp.Stop()
		return err
	}

	sp.SetText("Fetching weather data...")

	w, err := fetchWeather(location.Latitude, location.Longitude, units, forecast)

	sp.Stop()

	if err != nil {
		return err
	}

	render.CurrentWeather(out, location, w.Current, units, plain)

	if forecast && w.Daily != nil {
		render.Forecast(out, w.Daily, units, plain)
	}

	return nil
}

func printError(w io.Writer, err error) {
	fmt.Fprintln(w, render.Red("Error: "+err.Error()))
}
