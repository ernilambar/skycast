import boxen from 'boxen'
import chalk from 'chalk'

const WEATHER_ICONS = {
  Clear: chalk.yellow(`
    \\   /
     .-.
  ― (   ) ―
     \`-'
    /   \\     `),
  Rain: chalk.cyan(`
     .--.
    (    ).
   (___.__)
    ʻ ʻ ʻ ʻ
   ʻ ʻ ʻ ʻ    `),
  Clouds: chalk.gray(`
      .--.
   .-(    ).
  (___.__)__)
              `),
  Snow: chalk.white(`
     .--.
    (    ).
   (___.__)
    *  *  *
   *  *  *    `),
  Thunderstorm: chalk.yellow(`
     .--.
    (    ).
   (___.__)
   ⚡  ʻ  ⚡
    ʻ ʻ ʻ ʻ    `),
  Fog: chalk.gray(`
   _ - _ - _
  _ - _ - _ -
   _ - _ - _
              `)
}

const CONDITION_EMOJI = {
  Clear: '☀️',
  Clouds: '☁️',
  Rain: '🌧️',
  Snow: '❄️',
  Thunderstorm: '⛈️',
  Fog: '🌫️'
}

const SNOW_CODES = new Set([71, 73, 75, 77, 85, 86])
const RAIN_CODES = new Set([
  51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82
])

export function codeToCondition (code) {
  if (code === 0 || code === 1) return 'Clear'
  if (code === 2 || code === 3) return 'Clouds'
  if (code === 45 || code === 48) return 'Fog'
  if (code >= 95) return 'Thunderstorm'
  if (SNOW_CODES.has(code)) return 'Snow'
  if (RAIN_CODES.has(code)) return 'Rain'
  return 'Clouds'
}

function unitLabel (units) {
  return units === 'imperial' ? '°F' : '°C'
}

function formatTemp (value, units) {
  const text = `${Math.round(value)}${unitLabel(units)}`
  const celsius = units === 'imperial' ? ((value - 32) * 5) / 9 : value

  if (celsius <= 0) return chalk.cyan.bold(`${text} ❄️`)
  if (celsius <= 18) return chalk.blue.bold(`${text} 🌤️`)
  if (celsius <= 28) return chalk.yellow.bold(`${text} ☀️`)
  return chalk.red.bold(`${text} 🔥`)
}

function windUnitLabel (units) {
  return units === 'imperial' ? 'mph' : 'km/h'
}

function locationLabel (location) {
  return location.country
    ? `${location.city}, ${location.country}`
    : location.city
}

export function renderCurrentWeather (location, current, units, plain) {
  const condition = codeToCondition(current.code)

  if (plain) {
    console.log(`Location: ${locationLabel(location)}`)
    console.log(`Condition: ${condition}`)
    console.log(
      `Temperature: ${Math.round(current.temperature)}${unitLabel(units)}`
    )
    console.log(
      `Feels Like: ${Math.round(current.feelsLike)}${unitLabel(units)}`
    )
    console.log(`Humidity: ${current.humidity}%`)
    console.log(`Wind Speed: ${current.wind} ${windUnitLabel(units)}`)
    return
  }

  const icon = WEATHER_ICONS[condition]
  const cityTitle = chalk.bold.cyan(
    `  📍 ${locationLabel(location).toUpperCase()}  `
  )

  const stats = `
${chalk.bold('Condition:')}   ${chalk.italic(condition)}
${chalk.bold('Temperature:')} ${formatTemp(current.temperature, units)}
${chalk.bold('Feels Like:')}  ${Math.round(current.feelsLike)}${unitLabel(units)}
${chalk.bold('Humidity:')}    ${current.humidity}%
${chalk.bold('Wind Speed:')}  ${current.wind} ${windUnitLabel(units)}
`

  const cardBody = `${icon}\n${stats}`

  console.log(
    boxen(cardBody, {
      title: cityTitle,
      titleAlignment: 'center',
      padding: 1,
      margin: 1,
      borderStyle: 'round',
      borderColor: 'cyan'
    })
  )
}

function weekdayName (date) {
  return new Date(`${date}T00:00:00`).toLocaleDateString('en-US', {
    weekday: 'short'
  })
}

export function renderForecast (daily, units, plain) {
  if (plain) {
    console.log('\n5-Day Forecast:')
    for (const day of daily) {
      const condition = codeToCondition(day.code)
      console.log(
        `${weekdayName(day.date)}: ${condition}, High ${Math.round(day.tempMax)}${unitLabel(units)}, Low ${Math.round(day.tempMin)}${unitLabel(units)}`
      )
    }
    return
  }

  const rows = daily
    .map((day) => {
      const condition = codeToCondition(day.code)
      const label = weekdayName(day.date).padEnd(4)
      const high = chalk.red(`${Math.round(day.tempMax)}°`.padStart(5))
      const low = chalk.blue(`${Math.round(day.tempMin)}°`.padStart(5))
      return `${chalk.bold(label)} ${CONDITION_EMOJI[condition]}  ${condition.padEnd(13)} ${high} / ${low}`
    })
    .join('\n')

  console.log(
    boxen(rows, {
      title: chalk.bold.cyan('5-Day Forecast'),
      titleAlignment: 'center',
      padding: 1,
      margin: { top: 0, bottom: 1, left: 1, right: 1 },
      borderStyle: 'round',
      borderColor: 'cyan'
    })
  )
}
