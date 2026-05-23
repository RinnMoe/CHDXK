import type { ApiKeyDTO } from "@/api/api-key"

interface MockApiKey {
  id: number
  name: string
  key: string
  role: string
  user_id: number
  created_at: string
  last_used_at: string | null
}

interface MockApiKeyState {
  keys: MockApiKey[]
}

interface PersistedMockApiKeyState {
  version: 1
  keys: MockApiKey[]
}

const MOCK_API_KEY_STORAGE_KEY = "jcourse:mock-api-key-state"

const initialApiKeys: MockApiKey[] = [
  {
    id: 1,
    name: "本地脚本",
    key: "jc_demo_user_00000000000000000000000000000001",
    role: "user",
    user_id: 1,
    created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(),
    last_used_at: null,
  },
]

function createMockApiKeyState(): MockApiKeyState {
  return { keys: [...initialApiKeys] }
}

function readStoredMockApiKeyState(): MockApiKeyState | null {
  if (typeof window === "undefined") return null

  const raw = window.localStorage.getItem(MOCK_API_KEY_STORAGE_KEY)
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<PersistedMockApiKeyState>
    if (parsed.version !== 1 || !Array.isArray(parsed.keys)) {
      return null
    }
    return { keys: parsed.keys }
  } catch {
    window.localStorage.removeItem(MOCK_API_KEY_STORAGE_KEY)
    return null
  }
}

function persistMockApiKeyState() {
  if (typeof window === "undefined") return

  const persisted: PersistedMockApiKeyState = {
    version: 1,
    keys: mockApiKeyState.keys,
  }
  window.localStorage.setItem(
    MOCK_API_KEY_STORAGE_KEY,
    JSON.stringify(persisted)
  )
}

const mockApiKeyState: MockApiKeyState =
  import.meta.hot?.data.mockApiKeyState ??
  readStoredMockApiKeyState() ??
  createMockApiKeyState()

if (import.meta.hot) {
  import.meta.hot.dispose((data) => {
    data.mockApiKeyState = mockApiKeyState
  })
}

export const mockApiKeys = mockApiKeyState.keys

function maskApiKey(key: string) {
  if (key.length <= 12) return "*".repeat(key.length)
  return `${key.slice(0, 7)}************${key.slice(-6)}`
}

export function toApiKeyDTO(key: MockApiKey, includeKey = false): ApiKeyDTO {
  return {
    id: key.id,
    name: key.name,
    key: includeKey ? key.key : undefined,
    key_masked: maskApiKey(key.key),
    role: key.role,
    user_id: key.user_id,
    created_at: key.created_at,
    last_used_at: key.last_used_at,
  }
}

export function listMockUserApiKeys(userID: number): ApiKeyDTO[] {
  return mockApiKeys
    .filter((key) => key.role === "user" && key.user_id === userID)
    .sort((a, b) => b.created_at.localeCompare(a.created_at) || b.id - a.id)
    .map((key) => toApiKeyDTO(key))
}

export function createMockUserApiKey(userID: number, name: string): ApiKeyDTO {
  const now = new Date().toISOString()
  const id = Math.max(0, ...mockApiKeys.map((key) => key.id)) + 1
  const key: MockApiKey = {
    id,
    name,
    key: `jc_mock_${userID}_${crypto.randomUUID().replaceAll("-", "")}`,
    role: "user",
    user_id: userID,
    created_at: now,
    last_used_at: null,
  }
  mockApiKeys.push(key)
  persistMockApiKeyState()
  return toApiKeyDTO(key, true)
}

export function deleteMockUserApiKey(userID: number, id: number): boolean {
  const index = mockApiKeys.findIndex(
    (key) => key.id === id && key.role === "user" && key.user_id === userID
  )
  if (index < 0) return false
  mockApiKeys.splice(index, 1)
  persistMockApiKeyState()
  return true
}
