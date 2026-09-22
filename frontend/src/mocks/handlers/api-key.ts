import { http, HttpResponse } from "msw"
import type { CreateApiKeyCommand } from "@/api/api-key"
import {
  createMockSystemApiKey,
  deleteMockSystemApiKey,
  listMockSystemApiKeys,
} from "@/mocks/fixtures/api-keys"
import { findUserByID, mockSession } from "@/mocks/fixtures/auth"
import { randomDelay } from "@/mocks/utils"

function requireAdmin() {
  const user = mockSession.userID ? findUserByID(mockSession.userID) : null
  return user?.is_admin() ?? false
}

export const apiKeyHandlers = [
  http.get("/api/admin/api-key/system", async () => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!requireAdmin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    return HttpResponse.json(listMockSystemApiKeys())
  }),

  http.post("/api/admin/api-key/system", async ({ request }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!requireAdmin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    const body = (await request.json()) as CreateApiKeyCommand
    const name = body.name?.trim()
    if (!name) {
      return HttpResponse.json(
        { error: "api key name is required" },
        { status: 400 }
      )
    }
    return HttpResponse.json(createMockSystemApiKey(name), { status: 201 })
  }),

  http.delete("/api/admin/api-key/system/:apiKeyID", async ({ params }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    if (!requireAdmin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    const id = String(params.apiKeyID ?? "")
    if (!/^\d+$/.test(id) || id === "0") {
      return HttpResponse.json({ error: "invalid api key id" }, { status: 400 })
    }
    if (!deleteMockSystemApiKey(id)) {
      return HttpResponse.json({ error: "api key not found" }, { status: 404 })
    }
    return new HttpResponse(null, { status: 204 })
  }),
]
