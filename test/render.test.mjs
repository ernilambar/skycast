import test from 'node:test'
import assert from 'node:assert/strict'
import { codeToCondition } from '../src/render.mjs'

test('codeToCondition maps WMO codes to conditions', () => {
  assert.equal(codeToCondition(0), 'Clear')
  assert.equal(codeToCondition(1), 'Clear')
  assert.equal(codeToCondition(2), 'Clouds')
  assert.equal(codeToCondition(3), 'Clouds')
  assert.equal(codeToCondition(45), 'Fog')
  assert.equal(codeToCondition(48), 'Fog')
  assert.equal(codeToCondition(63), 'Rain')
  assert.equal(codeToCondition(80), 'Rain')
  assert.equal(codeToCondition(82), 'Rain')
  assert.equal(codeToCondition(71), 'Snow')
  assert.equal(codeToCondition(85), 'Snow')
  assert.equal(codeToCondition(95), 'Thunderstorm')
  assert.equal(codeToCondition(99), 'Thunderstorm')
})

test('codeToCondition falls back to Clouds for unknown codes', () => {
  assert.equal(codeToCondition(-1), 'Clouds')
})
