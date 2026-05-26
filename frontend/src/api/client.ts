import type { ApiError } from "./types"
import { BASE_URL } from "./constants"

class HttpError extends Error {
  status: number
  body?: ApiError

  constructor(message: string, status: number, body?: ApiError) {
    super(message)
    this.name = "HttpError"
    this.status = status
    this.body = body
  }
}

export async function apiClient<T>(
  input: RequestInfo,
  init?: RequestInit
): Promise<T> {
  const method = requestMethod(input, init)
  const headers = new Headers(init?.headers)
  if (!headers.has("Content-Type"))
    headers.set("Content-Type", "application/json")
  if (requiresCsrfToken(method) && !headers.has("X-CSRF-Token")) {
    headers.set("X-CSRF-Token", await fetchCsrfToken())
  }

  const res = await fetch(input, {
    ...init,
    credentials: "include",
    headers,
  })

  if (!res.ok) {
    let errorBody: ApiError | undefined
    try {
      errorBody = await res.json()
    } catch {
      // ignore parse error
    }
    throw new HttpError(
      errorBody?.error ?? `Request failed: ${res.status}`,
      res.status,
      errorBody
    )
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json() as Promise<T>
}

function requestMethod(input: RequestInfo, init?: RequestInit) {
  return (
    init?.method ?? (input instanceof Request ? input.method : "GET")
  ).toUpperCase()
}

function requiresCsrfToken(method: string) {
  return method !== "GET" && method !== "HEAD" && method !== "OPTIONS"
}

async function fetchCsrfToken() {
  const res = await fetch(`${BASE_URL}/auth/csrf`, {
    credentials: "include",
  })
  const token = res.headers.get("X-CSRF-Token")
  if (!res.ok || !token) throw new Error("failed to fetch csrf token")
  return token
}

export { HttpError }
