import { http, HttpResponse } from "msw"
import type { CreateApiKeyCommand } from "@/api/api-key"
import {
  createMockUserApiKey,
  deleteMockUserApiKey,
  listMockUserApiKeys,
} from "@/mocks/fixtures/api-keys"
import { mockSession } from "@/mocks/fixtures/auth"
import { randomDelay } from "@/mocks/utils"

export const apiKeyHandlers = [
  http.get("/api/api-key/", async () => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    return HttpResponse.json(listMockUserApiKeys(mockSession.userID))
  }),

  http.post("/api/api-key/", async ({ request }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const body = (await request.json()) as CreateApiKeyCommand
    const name = body.name?.trim()
    if (!name) {
      return HttpResponse.json(
        { error: "api key name is required" },
        { status: 400 }
      )
    }
    return HttpResponse.json(createMockUserApiKey(mockSession.userID, name), {
      status: 201,
    })
  }),

  http.delete("/api/api-key/:apiKeyID", async ({ params }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const id = String(params.apiKeyID ?? "")
    if (!/^\d+$/.test(id) || id === "0") {
      return HttpResponse.json({ error: "invalid api key id" }, { status: 400 })
    }
    if (!deleteMockUserApiKey(mockSession.userID, id)) {
      return HttpResponse.json({ error: "api key not found" }, { status: 404 })
    }
    return new HttpResponse(null, { status: 204 })
  }),
]
