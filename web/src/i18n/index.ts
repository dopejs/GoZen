import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from './locales/en.json'
import zhCN from './locales/zh-CN.json'
import zhTW from './locales/zh-TW.json'
import es from './locales/es.json'
import ja from './locales/ja.json'
import ko from './locales/ko.json'

const savedLang = localStorage.getItem('gozen-lang') || navigator.language || 'en'

i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    'zh-CN': { translation: zhCN },
    'zh-TW': { translation: zhTW },
    es: { translation: es },
    ja: { translation: ja },
    ko: { translation: ko },
  },
  lng: savedLang,
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false,
  },
})

export default i18n
