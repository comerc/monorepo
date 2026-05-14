import React, { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { App as AntdApp, ConfigProvider, theme } from 'antd'
import ruRU from 'antd/locale/ru_RU'

export type ThemeMode = 'light' | 'auto' | 'dark'
type EffectiveTheme = 'light' | 'dark'

interface ThemeState {
  mode: ThemeMode
  effective: EffectiveTheme
  setMode: (mode: ThemeMode) => void
}

const THEME_KEY = 'theme-store'
const ThemeContext = createContext<ThemeState | null>(null)

function readInitialMode(): ThemeMode {
  const raw = localStorage.getItem(THEME_KEY)
  if (raw === 'light' || raw === 'auto' || raw === 'dark') return raw

  try {
    const parsed = JSON.parse(raw ?? '') as { state?: { mode?: ThemeMode } }
    const mode = parsed.state?.mode
    if (mode === 'light' || mode === 'auto' || mode === 'dark') return mode
  } catch {
    // localStorage может содержать старый формат, тогда берём default
  }

  return 'auto'
}

function getSystemTheme(): EffectiveTheme {
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(readInitialMode)
  const [systemTheme, setSystemTheme] = useState<EffectiveTheme>(getSystemTheme)
  const effective = mode === 'auto' ? systemTheme : mode

  useEffect(() => {
    const query = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (event: MediaQueryListEvent) => {
      setSystemTheme(event.matches ? 'dark' : 'light')
    }

    query.addEventListener('change', handleChange)
    return () => query.removeEventListener('change', handleChange)
  }, [])

  useEffect(() => {
    document.documentElement.classList.toggle('dark', effective === 'dark')
  }, [effective])

  const value = useMemo<ThemeState>(
    () => ({
      mode,
      effective,
      setMode: (nextMode) => {
        localStorage.setItem(THEME_KEY, nextMode)
        setModeState(nextMode)
      },
    }),
    [effective, mode],
  )

  return (
    <ThemeContext.Provider value={value}>
      <ConfigProvider
        locale={ruRU}
        theme={{
          algorithm: effective === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
          zeroRuntime: true,
        }}
      >
        <AntdApp>{children}</AntdApp>
      </ConfigProvider>
    </ThemeContext.Provider>
  )
}

export function useThemeMode() {
  const value = useContext(ThemeContext)
  if (!value) {
    throw new Error('useThemeMode must be used inside ThemeProvider')
  }
  return value
}
