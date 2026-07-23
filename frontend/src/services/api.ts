import axios, { type InternalAxiosRequestConfig } from 'axios'
import { clearStoredSession, getActiveRole, getActiveSchoolId } from './session'
import { getAccessToken, setAccessToken } from './accessToken'

interface RetryableRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
}

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api',
  // Required so the refresh_token httpOnly cookie is actually sent/received —
  // it's on a different mechanism entirely from the Authorization header below.
  withCredentials: true,
})

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  const activeSchoolId = getActiveSchoolId()
  const activeRole = getActiveRole()

  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  if (activeSchoolId) {
    config.headers.SchoolId = activeSchoolId
  }

  if (activeSchoolId && activeRole) {
    config.headers['Active-Role'] = activeRole
  }

  return config
})

// Endpoints that must never trigger a refresh-retry or the global
// logout-redirect on their own 401 — otherwise a failed login attempt would
// try to "refresh" (nonsensical), and /refresh-token's own 401 (no cookie,
// or an expired/reused one) would recurse into trying to refresh itself, or
// wrongly redirect an anonymous visitor whose boot-time silent-refresh
// attempt simply found no session.
const AUTH_ENDPOINTS = ['/login', '/register', '/refresh-token', '/logout']

function isAuthEndpoint(url?: string): boolean {
  if (!url) return false
  return AUTH_ENDPOINTS.some((path) => url.includes(path))
}

// Single-flight guard: if several requests 401 around the same time (e.g.
// a page firing a few parallel GETs right as the access token expires),
// they all await this one shared promise instead of each POSTing
// /refresh-token independently.
let refreshPromise: Promise<string | null> | null = null

export function refreshAccessToken(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = api
      .post<{ accessToken: string }>('/refresh-token')
      .then(({ data }) => {
        setAccessToken(data.accessToken)
        return data.accessToken
      })
      .catch(() => {
        setAccessToken(null)
        return null
      })
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

function forceLogout() {
  setAccessToken(null)
  clearStoredSession()
  if (window.location.pathname !== '/login') {
    window.location.assign('/login')
  }
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as RetryableRequestConfig | undefined
    const status = error.response?.status

    if (status !== 401 || !originalRequest || isAuthEndpoint(originalRequest.url)) {
      return Promise.reject(error)
    }

    if (originalRequest._retry) {
      // Already retried once after a refresh and still 401 — refresh
      // itself must have been stale or the session is genuinely over.
      forceLogout()
      return Promise.reject(error)
    }

    originalRequest._retry = true
    const newToken = await refreshAccessToken()
    if (!newToken) {
      forceLogout()
      return Promise.reject(error)
    }

    originalRequest.headers = originalRequest.headers ?? {}
    originalRequest.headers.Authorization = `Bearer ${newToken}`
    return api(originalRequest)
  },
)
