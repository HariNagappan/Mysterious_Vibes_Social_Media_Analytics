import axios, { AxiosError } from "axios";

import type { ApiErrorShape } from "@/types";

const TOKEN_KEY = "pulsegraph.token";

export class ApiError extends Error {
  code: string;

  constructor(code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.code = code;
  }
}

const baseURL = (import.meta.env.VITE_API_URL ?? "").trim();
const forceMocks =
  String(import.meta.env.VITE_USE_MOCKS ?? "").toLowerCase() === "true";

/** True when the app serves deterministic mock data instead of HTTP. */
export function isMockMode(): boolean {
  return forceMocks || baseURL.length === 0;
}

export function hasBackend(): boolean {
  return !isMockMode();
}

export function getStoredToken(): string | null {
  try {
    return window.localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export function storeToken(token: string | null): void {
  try {
    if (token) window.localStorage.setItem(TOKEN_KEY, token);
    else window.localStorage.removeItem(TOKEN_KEY);
  } catch {
    /* storage unavailable - session-only auth */
  }
}

export const http = axios.create({
  baseURL,
  timeout: 12_000,
  headers: { "Content-Type": "application/json" },
});

http.interceptors.request.use((config) => {
  const token = getStoredToken();
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: ApiErrorShape }>) => {
    const payload = error.response?.data?.error;
    if (payload) {
      return Promise.reject(new ApiError(payload.code, payload.message));
    }
    return Promise.reject(
      new ApiError("network_error", error.message || "Network request failed"),
    );
  },
);
