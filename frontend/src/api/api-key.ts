import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export interface ApiKeyDTO {
  id: string
  name: string
  key: string
  role: string
  user_id: number
  created_at: string
  last_used_at?: string | null
}

export interface CreateApiKeyCommand {
  name: string
}

export function listApiKeys(): Promise<ApiKeyDTO[]> {
  return apiClient<ApiKeyDTO[]>(`${BASE_URL}/api-key/`)
}

export function createApiKey(cmd: CreateApiKeyCommand): Promise<ApiKeyDTO> {
  return apiClient<ApiKeyDTO>(`${BASE_URL}/api-key/`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function deleteApiKey(id: string): Promise<void> {
  return apiClient<void>(`${BASE_URL}/api-key/${id}`, {
    method: "DELETE",
  })
}

export function listSystemApiKeys(): Promise<ApiKeyDTO[]> {
  return apiClient<ApiKeyDTO[]>(`${BASE_URL}/admin/api-key/system`)
}

export function createSystemApiKey(
  cmd: CreateApiKeyCommand
): Promise<ApiKeyDTO> {
  return apiClient<ApiKeyDTO>(`${BASE_URL}/admin/api-key/system`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function deleteSystemApiKey(id: string): Promise<void> {
  return apiClient<void>(`${BASE_URL}/admin/api-key/system/${id}`, {
    method: "DELETE",
  })
}
