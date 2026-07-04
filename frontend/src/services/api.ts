import axios from 'axios'
import { clearStoredSession, getActiveRole, getActiveSchoolId, getStoredToken } from './session'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api',
})

api.interceptors.request.use((config) => {
  const token = getStoredToken()
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

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && window.location.pathname !== '/login') {
      clearStoredSession()
      window.location.assign('/login')
    }
    return Promise.reject(error)
  },
)
