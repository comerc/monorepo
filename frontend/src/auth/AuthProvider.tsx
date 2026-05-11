import { createContext, useContext, useMemo, useState } from 'react'

interface AuthUser {
  id: string
  email: string
  nickname?: string | null
}

interface AuthState {
  token: string | null
  user: AuthUser | null
  isAuthenticated: boolean
  login: (token: string, user: AuthUser) => void
  updateUser: (user: AuthUser) => void
  logout: () => void
}

const TOKEN_KEY = 'auth_token'
const USER_KEY = 'auth_user'

const AuthContext = createContext<AuthState | null>(null)

function loadUser(): AuthUser | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null

  try {
    return JSON.parse(raw) as AuthUser
  } catch {
    localStorage.removeItem(USER_KEY)
    return null
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<AuthUser | null>(() => loadUser())

  const value = useMemo<AuthState>(
    () => ({
      token,
      user,
      isAuthenticated: Boolean(token),
      login: (nextToken, nextUser) => {
        localStorage.setItem(TOKEN_KEY, nextToken)
        localStorage.setItem(USER_KEY, JSON.stringify(nextUser))
        setToken(nextToken)
        setUser(nextUser)
      },
      updateUser: (nextUser) => {
        localStorage.setItem(USER_KEY, JSON.stringify(nextUser))
        setUser(nextUser)
      },
      logout: () => {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_KEY)
        setToken(null)
        setUser(null)
      },
    }),
    [token, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) {
    throw new Error('useAuth must be used inside AuthProvider')
  }
  return value
}
