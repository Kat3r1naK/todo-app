import { defineStore } from 'pinia'

const THEME_KEY = 'todo.theme'

export type ThemeId = 'night' | 'sunrise' | 'mint' | 'sakura' | 'paper'

export interface ThemeOption {
  id: ThemeId
  label: string
}

export const THEMES: ThemeOption[] = [
  { id: 'paper', label: '纸白' },
  { id: 'night', label: '石墨' },
  { id: 'sunrise', label: '暖沙' },
  { id: 'mint', label: '薄荷' },
  { id: 'sakura', label: '浅樱' },
]

function normalizeTheme(value: string | null): ThemeId {
  return THEMES.some((item) => item.id === value) ? (value as ThemeId) : 'paper'
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
