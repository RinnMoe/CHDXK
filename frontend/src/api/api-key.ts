import { apiClient } from "./client"

const BASE = "/api/api-keys"

export interface ApiKeyDTO {
  id: number
  name: string
  key?: string
  key_masked: string
  role: string
  user_id: number
  created_at: string
  last_used_at?: string | null
}

export interface CreateApiKeyCommand {
  name: string
}

export function listApiKeys(): Promise<ApiKeyDTO[]> {
  return apiClient<ApiKeyDTO[]>(`${BASE}/`)
}

export function createApiKey(cmd: CreateApiKeyCommand): Promise<ApiKeyDTO> {
  return apiClient<ApiKeyDTO>(`${BASE}/`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function deleteApiKey(id: number): Promise<void> {
  return apiClient<void>(`${BASE}/${id}`, {
    method: "DELETE",
  })
}
