import type { ApiKeyDTO } from "@/api/api-key"

interface MockApiKey {
  id: string
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
    id: "1",
    name: "本地脚本",
    key: mockApiKeyValue("1"),
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
  const [prefix, keyID] = key.split("_")
  return `${prefix}_${keyID}_********`
}

function mockApiKeyValue(id: string) {
  return `jc_${mockEncodedKeyID(id)}_${mockSecret()}`
}

function mockEncodedKeyID(id: string) {
  let value = BigInt(id)
  const bytes = new Uint8Array(8)
  for (let i = bytes.length - 1; i >= 0; i -= 1) {
    bytes[i] = Number(value & 0xffn)
    value >>= 8n
  }
  return hex(bytes)
}

function mockSecret() {
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  return hex(bytes)
}

function hex(bytes: Uint8Array) {
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join(
    ""
  )
}

export function toApiKeyDTO(key: MockApiKey, includeKey = false): ApiKeyDTO {
  return {
    id: key.id,
    name: key.name,
    key: includeKey ? key.key : maskApiKey(key.key),
    role: key.role,
    user_id: key.user_id,
    created_at: key.created_at,
    last_used_at: key.last_used_at,
  }
}

export function listMockUserApiKeys(userID: number): ApiKeyDTO[] {
  return mockApiKeys
    .filter((key) => key.role === "user" && key.user_id === userID)
    .sort(
      (a, b) =>
        b.created_at.localeCompare(a.created_at) || Number(b.id) - Number(a.id)
    )
    .map((key) => toApiKeyDTO(key))
}

export function createMockUserApiKey(userID: number, name: string): ApiKeyDTO {
  const now = new Date().toISOString()
  const id = String(
    Math.max(0, ...mockApiKeys.map((key) => Number(key.id))) + 1
  )
  const key: MockApiKey = {
    id,
    name,
    key: mockApiKeyValue(id),
    role: "user",
    user_id: userID,
    created_at: now,
    last_used_at: null,
  }
  mockApiKeys.push(key)
  persistMockApiKeyState()
  return toApiKeyDTO(key, true)
}

export function deleteMockUserApiKey(userID: number, id: string): boolean {
  const index = mockApiKeys.findIndex(
    (key) => key.id === id && key.role === "user" && key.user_id === userID
  )
  if (index < 0) return false
  mockApiKeys.splice(index, 1)
  persistMockApiKeyState()
  return true
}
