import { http, HttpResponse } from "msw"
import {
  addUser,
  consumeMockCode,
  findUserByEmail,
  findUserByID,
  mockSession,
  setMockCode,
  setMockSessionUserID,
  setMockUserPassword,
  toAuthUserDTO,
} from "@/mocks/fixtures/auth"
import { getMockSystemSettingValue } from "./system-settings"
import { randomDelay } from "@/mocks/utils"

function parseStringList(value: string | undefined) {
  return (value ?? "@sjtu.edu.cn")
    .split(/[\s,]+/)
    .map((item) => item.trim().toLowerCase())
    .filter(Boolean)
}

function isAllowedRegistrationEmail(email: string) {
  const allowed = parseStringList(
    getMockSystemSettingValue("auth.registration.email_whitelist")
  )
  return allowed.some((item) =>
    item.startsWith("@") ? email.endsWith(item) : email === item
  )
}

export const authHandlers = [
  http.post("/api/auth/register/code", async ({ request }) => {
    await randomDelay()
    const body = (await request.json()) as { email: string }
    if (!body.email) {
      return HttpResponse.json({ error: "email is required" }, { status: 400 })
    }
    const email = body.email.toLowerCase()
    if (!isAllowedRegistrationEmail(email)) {
      return HttpResponse.json(
        { error: "email domain not allowed" },
        { status: 403 }
      )
    }
    if (findUserByEmail(email)) {
      return HttpResponse.json(
        { error: "user already exists" },
        { status: 409 }
      )
    }
    setMockCode(email)
    return HttpResponse.json({ message: "ok" })
  }),

  http.post("/api/auth/register", async ({ request }) => {
    await randomDelay()
    const body = (await request.json()) as {
      email: string
      code: string
      password: string
    }
    if (!body.email || !body.code || !body.password) {
      return HttpResponse.json(
        { error: "missing required fields" },
        { status: 400 }
      )
    }
    if (!consumeMockCode(body.email, body.code)) {
      return HttpResponse.json(
        { error: "invalid verification code" },
        { status: 400 }
      )
    }
    const user = addUser(body.email, body.password)
    setMockSessionUserID(user.id)
    return HttpResponse.json(toAuthUserDTO(user), { status: 201 })
  }),

  http.post("/api/auth/login", async ({ request }) => {
    await randomDelay()
    const body = (await request.json()) as { email: string; password: string }
    if (!body.email || !body.password) {
      return HttpResponse.json(
        { error: "missing required fields" },
        { status: 400 }
      )
    }
    const user = findUserByEmail(body.email)
    if (!user || user.password !== body.password) {
      return HttpResponse.json(
        { error: "invalid credentials" },
        { status: 401 }
      )
    }
    setMockSessionUserID(user.id)
    return HttpResponse.json(toAuthUserDTO(user))
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

  http.post("/api/auth/password-reset/code", async ({ request }) => {
    await randomDelay()
    const body = (await request.json()) as { email: string }
    if (!body.email) {
      return HttpResponse.json({ error: "email is required" }, { status: 400 })
    }
    if (!findUserByEmail(body.email)) {
      return HttpResponse.json({ error: "user not found" }, { status: 404 })
    }
    setMockCode(body.email)
    return HttpResponse.json({ message: "ok" })
  }),

  http.post("/api/auth/password-reset", async ({ request }) => {
    await randomDelay()
    const body = (await request.json()) as {
      email: string
      code: string
      new_password: string
    }
    if (!body.email || !body.code || !body.new_password) {
      return HttpResponse.json(
        { error: "missing required fields" },
        { status: 400 }
      )
    }
    if (!consumeMockCode(body.email, body.code)) {
      return HttpResponse.json(
        { error: "invalid verification code" },
        { status: 400 }
      )
    }
    const user = findUserByEmail(body.email)
    if (!user) {
      return HttpResponse.json({ error: "user not found" }, { status: 404 })
    }
    setMockUserPassword(user, body.new_password)
    return HttpResponse.json({ message: "ok" })
  }),
]
