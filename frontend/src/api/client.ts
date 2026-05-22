import type { ApiError } from "./types"

class HttpError extends Error {
  constructor(
    message: string,
    public status: number,
    public body?: ApiError
  ) {
    super(message)
    this.name = "HttpError"
  }
}

export async function apiClient<T>(
  input: RequestInfo,
  init?: RequestInit
): Promise<T> {
  const res = await fetch(input, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
    ...init,
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

export { HttpError }
