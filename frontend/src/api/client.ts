import { copy, translateMessage } from "@/lib/copy"
import type { ApiError } from "./types"
import { BASE_URL } from "./constants"

const CSRF_HEADER = "X-CSRF-Token"
const CSRF_ERRORS = new Set(["csrf token missing", "csrf token mismatch"])

const protocolMessages: Record<string, string> = {
  unauthorized: copy("copy.user.api.client.ab1e950a6480"),
  forbidden: copy("copy.user.api.client.10f0e81879c6"),
  "internal error": copy("copy.user.api.client.5b1606fabf1e"),
  "csrf token missing": copy("copy.user.api.client.f8194185f36b"),
  "csrf token mismatch": copy("copy.user.api.client.f8194185f36b"),
  "rate limit exceeded": copy("copy.user.api.client.e0f555394c57"),
  "invalid user id": copy("copy.user.api.client.c6dae0b087bc"),
}

let csrfToken: string | null = null
let csrfTokenRequest: Promise<string> | null = null

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
  const requiresCsrf = requiresCsrfToken(method)
  const hasCallerCsrfToken = new Headers(init?.headers).has(CSRF_HEADER)

  let res = await fetchWithDefaults(
    input,
    init,
    requiresCsrf && !hasCallerCsrfToken ? await fetchCsrfToken() : undefined
  )

  if (requiresCsrf && !hasCallerCsrfToken && (await isCsrfFailure(res))) {
    clearCsrfToken()
    res = await fetchWithDefaults(input, init, await fetchCsrfToken())
  }

  return parseResponse<T>(res)
}

async function fetchWithDefaults(
  input: RequestInfo,
  init?: RequestInit,
  csrf?: string
) {
  const headers = new Headers(init?.headers)
  const multipartBody =
    typeof FormData !== "undefined" && init?.body instanceof FormData
  if (multipartBody) headers.delete("Content-Type")
  else if (!headers.has("Content-Type"))
    headers.set("Content-Type", "application/json")
  if (csrf && !headers.has(CSRF_HEADER)) headers.set(CSRF_HEADER, csrf)

  return fetch(input, {
    ...init,
    credentials: "include",
    headers,
  })
}

async function parseResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorBody: ApiError | undefined
    try {
      errorBody = await res.json()
    } catch {
      // ignore parse error
    }
    throw new HttpError(
      translateMessage(
        protocolMessages[errorBody?.error ?? ""] ??
          errorBody?.error ??
          copy("copy.user.api.client.7f83dc051d39")
      ),
      res.status,
      errorBody
    )
  }

  if (res.status === 204) {
    return undefined as T
  }

  return translatePlaceholders(await res.json()) as T
}

function translatePlaceholders(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(translatePlaceholders)
  if (value && typeof value === "object") {
    const result = value as Record<string, unknown>
    const teacher = result.main_teacher as
      | { id?: number; name?: string }
      | undefined
    if (teacher?.id === 0)
      teacher.name = translateMessage(copy("copy.user.api.client.521ea9788171"))
    for (const key of Object.keys(result))
      result[key] = translatePlaceholders(result[key])
  }
  return value
}

async function isCsrfFailure(res: Response) {
  if (res.status !== 403) return false
  try {
    const body = (await res.clone().json()) as ApiError
    return CSRF_ERRORS.has(body.error)
  } catch {
    return false
  }
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
  if (csrfToken) return csrfToken
  if (!csrfTokenRequest) {
    csrfTokenRequest = fetch(`${BASE_URL}/auth/csrf`, {
      credentials: "include",
    })
      .then((res) => {
        const token = res.headers.get(CSRF_HEADER)
        if (!res.ok || !token)
          throw new Error(
            translateMessage(copy("copy.user.api.client.f8194185f36b"))
          )
        csrfToken = token
        return token
      })
      .finally(() => {
        csrfTokenRequest = null
      })
  }
  return csrfTokenRequest
}

function clearCsrfToken() {
  csrfToken = null
}

export { HttpError }
