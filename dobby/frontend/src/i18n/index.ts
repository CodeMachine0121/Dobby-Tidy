import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import zhTW from './locales/zh-TW'
import en from './locales/en'

const SUPPORTED_LANGS = ['zh-TW', 'en'] as const
export type SupportedLang = (typeof SUPPORTED_LANGS)[number]

export function resolveInitialLanguage(stored: string | null): SupportedLang {
  return stored && (SUPPORTED_LANGS as readonly string[]).includes(stored)
    ? (stored as SupportedLang)
    : 'zh-TW'
}

const initialLang = resolveInitialLanguage(localStorage.getItem('language'))

i18n.use(initReactI18next).init({
  resources: {
    'zh-TW': { translation: zhTW },
    en: { translation: en },
  },
  lng: initialLang,
  fallbackLng: 'zh-TW',
  interpolation: { escapeValue: false },
})

export default i18n
