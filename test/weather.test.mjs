import test from 'node:test'
import assert from 'node:assert/strict'
import { fetchWeather, geocodeCity, locateByIp } from '../src/weather.mjs'

function mockFetch (body, ok = true) {
  globalThis.fetch = async () => ({
    ok,
    json: async () => body
  })
}

test('geocodeCity returns the first matching result', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({
    results: [
      {
        name: 'Tokyo',
        country: 'Japan',
        latitude: 35.6895,
        longitude: 139.6917
      }
    ]
  })

  const location = await geocodeCity('Tokyo')

  assert.deepEqual(location, {
    city: 'Tokyo',
    country: 'Japan',
    latitude: 35.6895,
    longitude: 139.6917
  })
})

test('geocodeCity throws when no results are found', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({ results: [] })

  await assert.rejects(() => geocodeCity('Nowhereville'), /not found/)
})

test('locateByIp returns location from IP lookup', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({
    city: 'Kathmandu',
    country_name: 'Nepal',
    latitude: 27.7172,
    longitude: 85.324
  })

  const location = await locateByIp()

  assert.deepEqual(location, {
    city: 'Kathmandu',
    country: 'Nepal',
    latitude: 27.7172,
    longitude: 85.324
  })
})

test('locateByIp throws when the IP cannot be resolved', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({ error: true, reason: 'RateLimited' })

  await assert.rejects(() => locateByIp(), /Unable to detect location/)
})

test('fetchWeather parses current weather', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({
    current: {
      temperature_2m: 21,
      relative_humidity_2m: 60,
      apparent_temperature: 23,
      weather_code: 3,
      wind_speed_10m: 5
    }
  })

  const weather = await fetchWeather(27.7, 85.3, 'metric', false)

  assert.deepEqual(weather, {
    current: {
      temperature: 21,
      feelsLike: 23,
      humidity: 60,
      wind: 5,
      code: 3
    }
  })
})

test('fetchWeather parses the daily forecast when requested', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({
    current: {
      temperature_2m: 21,
      relative_humidity_2m: 60,
      apparent_temperature: 23,
      weather_code: 3,
      wind_speed_10m: 5
    },
    daily: {
      time: ['2026-08-01', '2026-08-02'],
      weather_code: [3, 61],
      temperature_2m_max: [30, 28],
      temperature_2m_min: [20, 19]
    }
  })

  const weather = await fetchWeather(27.7, 85.3, 'metric', true)

  assert.deepEqual(weather.daily, [
    { date: '2026-08-01', code: 3, tempMax: 30, tempMin: 20 },
    { date: '2026-08-02', code: 61, tempMax: 28, tempMin: 19 }
  ])
})

test('fetchWeather throws when the API responds with an error status', async (t) => {
  t.after(() => {
    globalThis.fetch = fetch
  })

  mockFetch({}, false)

  await assert.rejects(
    () => fetchWeather(0, 0, 'metric', false),
    /weather service/
  )
})
