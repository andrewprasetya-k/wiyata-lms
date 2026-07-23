import axios, { isAxiosError, type InternalAxiosRequestConfig } from "axios";
import {
  clearStoredSession,
  getActiveRole,
  getActiveSchoolId,
} from "./session";
import { getAccessToken, setAccessToken } from "./accessToken";

interface RetryableRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api",

  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const token = getAccessToken();
  const activeSchoolId = getActiveSchoolId();
  const activeRole = getActiveRole();

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  if (activeSchoolId) {
    config.headers.SchoolId = activeSchoolId;
  }

  if (activeSchoolId && activeRole) {
    config.headers["Active-Role"] = activeRole;
  }

  return config;
});

const AUTH_ENDPOINTS = ["/login", "/register", "/refresh-token", "/logout"];

function isAuthEndpoint(url?: string): boolean {
  if (!url) return false;
  return AUTH_ENDPOINTS.some((path) => url.includes(path));
}

const REFRESH_RETRY_DELAYS_MS = [1000, 2000];

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

async function requestNewAccessToken(): Promise<string | null> {
  for (let attempt = 0; ; attempt++) {
    try {
      const { data } = await api.post<{ accessToken: string }>(
        "/refresh-token",
      );
      setAccessToken(data.accessToken);
      return data.accessToken;
    } catch (error) {
      const status = isAxiosError(error) ? error.response?.status : undefined;
      if (status === 429 && attempt < REFRESH_RETRY_DELAYS_MS.length) {
        await delay(REFRESH_RETRY_DELAYS_MS[attempt]);
        continue;
      }
      // 401 (invalid/expired/reused), or 429 with retries exhausted, or
      // any other failure — nothing left to do but report failure.
      setAccessToken(null);
      return null;
    }
  }
}

let refreshPromise: Promise<string | null> | null = null;

export function refreshAccessToken(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = requestNewAccessToken().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}

function forceLogout() {
  setAccessToken(null);
  clearStoredSession();
  if (window.location.pathname !== "/login") {
    window.location.assign("/login");
  }
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as RetryableRequestConfig | undefined;
    const status = error.response?.status;

    if (
      status !== 401 ||
      !originalRequest ||
      isAuthEndpoint(originalRequest.url)
    ) {
      return Promise.reject(error);
    }

    if (originalRequest._retry) {
      forceLogout();
      return Promise.reject(error);
    }

    originalRequest._retry = true;
    const newToken = await refreshAccessToken();
    if (!newToken) {
      forceLogout();
      return Promise.reject(error);
    }

    originalRequest.headers = originalRequest.headers ?? {};
    originalRequest.headers.Authorization = `Bearer ${newToken}`;
    return api(originalRequest);
  },
);
