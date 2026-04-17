/**
 * Unit tests for i18n initialization logic
 *
 * Coverage:
 *   S-1  Default language is zh-TW when no localStorage entry
 *   S-2  Language persists from localStorage after restart
 *   S-3  Falls back to zh-TW when localStorage contains unsupported language
 *   S-15 zh-TW and en locale files have identical top-level keys
 *
 * Not covered (UI / React rendering):
 *   S-4 to S-14  — require browser + React Testing Library
 */

import { describe, it, expect } from 'vitest'
import { resolveInitialLanguage } from '../i18n/index'

// ── S-1: Default language ───────────────────────────────────────────────────────

describe('resolveInitialLanguage', () => {
  // S-1-1
  it('returns zh-TW when localStorage has no language entry', () => {
    // Arrange
    const localStorageValue = null

    // Act
    const result = resolveInitialLanguage(localStorageValue)

    // Assert
    expect(result).toBe('zh-TW')
  })

  // S-2-1
  it('returns "en" when localStorage stores "en"', () => {
    // Arrange
    const localStorageValue = 'en'

    // Act
    const result = resolveInitialLanguage(localStorageValue)

    // Assert
    expect(result).toBe('en')
  })

  // S-2-2
  it('returns "zh-TW" when localStorage stores "zh-TW"', () => {
    // Arrange
    const localStorageValue = 'zh-TW'

    // Act
    const result = resolveInitialLanguage(localStorageValue)

    // Assert
    expect(result).toBe('zh-TW')
  })

  // S-3-1
  it('falls back to zh-TW when localStorage stores an unsupported language code', () => {
    // Arrange
    const localStorageValue = 'fr'

    // Act
    const result = resolveInitialLanguage(localStorageValue)

    // Assert
    expect(result).toBe('zh-TW')
  })

  // S-3-2
  it('falls back to zh-TW when localStorage stores an empty string', () => {
    // Arrange
    const localStorageValue = ''

    // Act
    const result = resolveInitialLanguage(localStorageValue)

    // Assert
    expect(result).toBe('zh-TW')
  })
})

// ── S-15: Locale key alignment ────────────────────────────────────────────────

describe('locale file key alignment', () => {
  // S-15-1
  it('zh-TW and en have identical top-level keys', async () => {
    // Arrange
    const [zhTW, en] = await Promise.all([
      import('../i18n/locales/zh-TW').then((m) => m.default),
      import('../i18n/locales/en').then((m) => m.default),
    ])

    // Act
    const zhKeys = Object.keys(zhTW).sort()
    const enKeys = Object.keys(en).sort()

    // Assert
    expect(zhKeys).toEqual(enKeys)
  })

  // S-15-2
  it('zh-TW and en have identical nested keys under each section', async () => {
    // Arrange
    const [zhTW, en] = await Promise.all([
      import('../i18n/locales/zh-TW').then((m) => m.default),
      import('../i18n/locales/en').then((m) => m.default),
    ])

    function collectKeys(obj: Record<string, unknown>, prefix = ''): string[] {
      return Object.entries(obj).flatMap(([k, v]) => {
        const full = prefix ? `${prefix}.${k}` : k
        return typeof v === 'object' && v !== null
          ? collectKeys(v as Record<string, unknown>, full)
          : [full]
      })
    }

    // Act
    const zhKeys = collectKeys(zhTW as unknown as Record<string, unknown>).sort()
    const enKeys = collectKeys(en as unknown as Record<string, unknown>).sort()

    // Assert
    expect(zhKeys).toEqual(enKeys)
  })
})
