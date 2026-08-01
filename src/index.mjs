#!/usr/bin/env bun
import chalk from 'chalk'
import ora from 'ora'
import yargs from 'yargs'
import { hideBin } from 'yargs/helpers'
import { fetchWeather, geocodeCity, locateByIp } from './weather.mjs'
import { renderCurrentWeather, renderForecast } from './render.mjs'

async function main (city, options) {
  const spinner = ora()

  try {
    spinner.start(
      city ? `Looking up ${city}...` : 'Detecting your location...'
    )
    const location = city ? await geocodeCity(city) : await locateByIp()

    spinner.text = 'Fetching weather data...'
    const weather = await fetchWeather(
      location.latitude,
      location.longitude,
      options.units,
      options.forecast
    )

    spinner.stop()

    renderCurrentWeather(
      location,
      weather.current,
      options.units,
      options.plain
    )

    if (options.forecast && weather.daily) {
      renderForecast(weather.daily, options.units, options.plain)
    }
  } catch (error) {
    spinner.stop()
    console.error(
      chalk.red(
        `Error: ${error instanceof Error ? error.message : String(error)}`
      )
    )
    process.exit(1)
  }
}

const argv = await yargs(hideBin(process.argv))
  .scriptName('skycast')
  .usage('$0 [city]', 'A terminal weather app')
  .command('$0 [city]', 'Fetch weather for a city', (y) =>
    y.positional('city', {
      describe:
        'City name to fetch weather for (defaults to IP location)',
      type: 'string'
    })
  )
  .option('forecast', {
    alias: 'f',
    type: 'boolean',
    describe: 'Show 5-day forecast',
    default: false
  })
  .option('units', {
    alias: 'u',
    type: 'string',
    choices: ['metric', 'imperial'],
    describe: 'Temperature units: metric (C) or imperial (F)',
    default: 'metric'
  })
  .option('plain', {
    alias: 'p',
    type: 'boolean',
    describe: 'Output simple plain text without ASCII borders',
    default: false
  })
  .version('1.0.0')
  .alias('version', 'v')
  .help()
  .alias('help', 'h')
  .strict()
  .parse()

await main(argv.city, {
  forecast: argv.forecast,
  units: argv.units,
  plain: argv.plain
})
