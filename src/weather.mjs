export async function geocodeCity (city) {
  const url = `https://geocoding-api.open-meteo.com/v1/search?name=${encodeURIComponent(city)}&count=1`
  const res = await fetch(url)

  if (!res.ok) {
    throw new Error('Unable to reach the geocoding service')
  }

  const data = await res.json()
  const result = data.results?.[0]

  if (!result) {
    throw new Error(`City "${city}" not found`)
  }

  return {
    city: result.name,
    country: result.country ?? '',
    latitude: result.latitude,
    longitude: result.longitude
  }
}

export async function locateByIp () {
  const res = await fetch('https://ipapi.co/json/')

  if (!res.ok) {
    throw new Error('Unable to reach the IP location service')
  }

  const data = await res.json()

  if (data.error || !data.latitude) {
    throw new Error('Unable to detect location from IP address')
  }

  return {
    city: data.city,
    country: data.country_name ?? '',
    latitude: data.latitude,
    longitude: data.longitude
  }
}

export async function fetchWeather (latitude, longitude, units, forecast) {
  const temperatureUnit = units === 'imperial' ? 'fahrenheit' : 'celsius'
  const windSpeedUnit = units === 'imperial' ? 'mph' : 'kmh'

  const params = new URLSearchParams({
    latitude: String(latitude),
    longitude: String(longitude),
    current:
      'temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m',
    temperature_unit: temperatureUnit,
    wind_speed_unit: windSpeedUnit,
    timezone: 'auto'
  })

  if (forecast) {
    params.set(
      'daily',
      'weather_code,temperature_2m_max,temperature_2m_min'
    )
    params.set('forecast_days', '5')
  }

  const res = await fetch(
    `https://api.open-meteo.com/v1/forecast?${params.toString()}`
  )

  if (!res.ok) {
    throw new Error('Unable to reach the weather service')
  }

  const data = await res.json()

  const weather = {
    current: {
      temperature: data.current.temperature_2m,
      feelsLike: data.current.apparent_temperature,
      humidity: data.current.relative_humidity_2m,
      wind: data.current.wind_speed_10m,
      code: data.current.weather_code
    }
  }

  if (forecast && data.daily) {
    weather.daily = data.daily.time.map((date, i) => ({
      date,
      code: data.daily.weather_code[i],
      tempMax: data.daily.temperature_2m_max[i],
      tempMin: data.daily.temperature_2m_min[i]
    }))
  }

  return weather
}
