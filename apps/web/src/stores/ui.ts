import { defineStore } from 'pinia'

const THEME_KEY = 'todo.theme'

export type ThemeId = 'night' | 'sunrise' | 'mint' | 'sakura' | 'paper'

export interface ThemeOption {
  id: ThemeId
  label: string
}

export const THEMES: ThemeOption[] = [
  { id: 'night', label: '夜幕蓝' },
  { id: 'sunrise', label: '晨曦暖阳' },
  { id: 'mint', label: '薄荷清新' },
  { id: 'sakura', label: '樱花粉' },
  { id: 'paper', label: '纸墨米' },
]

function normalizeTheme(value: string | null): ThemeId {
  return THEMES.some((item) => item.id === value) ? (value as ThemeId) : 'night'
}

interface UiState {
  theme: ThemeId
}

export const useUiStore = defineStore('ui', {
  state: (): UiState => ({
    theme: normalizeTheme(localStorage.getItem(THEME_KEY)),
  }),
  actions: {
    init(): void {
      this.apply(this.theme)
    },
    apply(theme: string): void {
      const valid = normalizeTheme(theme)
      this.theme = valid
      localStorage.setItem(THEME_KEY, valid)
      const root = document.documentElement
      root.setAttribute('data-theme', valid)
      root.classList.toggle('dark', valid === 'night')
    },
  },
})
