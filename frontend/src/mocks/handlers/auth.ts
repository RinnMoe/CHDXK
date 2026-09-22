import { http, HttpResponse } from "msw"
import {
  findUserByID,
  mockSession,
  setMockSessionUserID,
  toAuthUserDTO,
} from "@/mocks/fixtures/auth"
import { randomDelay } from "@/mocks/utils"

export const authHandlers = [
  // Development-only stand-in for the browser redirect. The real backend
  // performs OIDC state, PKCE, nonce and token validation before returning.
  http.get("/api/auth/oidc/start", async ({ request }) => {
    await randomDelay()
    const requested = new URL(request.url).searchParams.get("return_to") ?? "/"
    const returnTo = requested.startsWith("/") && !requested.startsWith("//")
      ? requested
      : "/"
    setMockSessionUserID(1)
    return HttpResponse.redirect(new URL(returnTo, request.url).toString())
  }),

  http.post("/api/auth/logout", async () => {
    await randomDelay()
    setMockSessionUserID(null)
    return HttpResponse.json({ message: "ok" })
  }),

  http.get("/api/auth/me", async () => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const user = findUserByID(mockSession.userID)
    if (!user) {
      setMockSessionUserID(null)
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    return HttpResponse.json(toAuthUserDTO(user))
  }),
]
