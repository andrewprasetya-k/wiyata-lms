import axios, { isAxiosError, type InternalAxiosRequestConfig } from 'axios'
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

// A 429 from /refresh-token is transient (e.g. a burst of parallel requests
// during a context switch briefly exhausting the IP-scoped rate limit on
// the backend) — worth a couple of short retries before giving up. A 401
// means the token is genuinely invalid/expired/reused; retrying won't help,
// so it fails immediately. Increasing delays, capped at 2 retries: enough
// to ride out a brief burst without making a real session failure feel slow.
const REFRESH_RETRY_DELAYS_MS = [1000, 2000]

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function requestNewAccessToken(): Promise<string | null> {
  for (let attempt = 0; ; attempt++) {
    try {
      const { data } = await api.post<{ accessToken: string }>('/refresh-token')
      setAccessToken(data.accessToken)
      return data.accessToken
    } catch (error) {
      const status = isAxiosError(error) ? error.response?.status : undefined
      if (status === 429 && attempt < REFRESH_RETRY_DELAYS_MS.length) {
        await delay(REFRESH_RETRY_DELAYS_MS[attempt])
        continue
      }
      // 401 (invalid/expired/reused), or 429 with retries exhausted, or
      // any other failure — nothing left to do but report failure.
      setAccessToken(null)
      return null
    }
  }
}

// Single-flight guard: if several requests 401 around the same time (e.g.
// a page firing a few parallel GETs right as the access token expires),
// they all await this one shared promise instead of each POSTing
// /refresh-token independently. The retry-with-backoff above happens
// entirely inside this one promise, so callers already awaiting it
// transparently receive the result of the retried attempt too — nothing
// spawns a second, parallel refresh call.
let refreshPromise: Promise<string | null> | null = null

export function refreshAccessToken(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = requestNewAccessToken().finally(() => {
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
