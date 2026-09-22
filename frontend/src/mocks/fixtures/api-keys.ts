import type { ApiKeyDTO } from "@/api/api-key"

interface MockSystemApiKey {
  id: string
  name: string
  key: string
  created_at: string
  last_used_at: string | null
}

interface PersistedMockSystemApiKeyState {
  version: 1
  keys: MockSystemApiKey[]
}

const MOCK_API_KEY_STORAGE_KEY = "jcourse:mock-system-api-key-state"

const initialApiKeys: MockSystemApiKey[] = [
  {
    id: "1",
    name: "本地脚本",
    key: mockApiKeyValue("1"),
    created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(),
    last_used_at: null,
  },
]

function readStoredKeys(): MockSystemApiKey[] | null {
  if (typeof window === "undefined") return null
  const raw = window.localStorage.getItem(MOCK_API_KEY_STORAGE_KEY)
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<PersistedMockSystemApiKeyState>
    if (parsed.version !== 1 || !Array.isArray(parsed.keys)) return null
    return parsed.keys
  } catch {
    window.localStorage.removeItem(MOCK_API_KEY_STORAGE_KEY)
    return null
  }
}

function persistKeys(keys: MockSystemApiKey[]) {
  if (typeof window === "undefined") return
  const value: PersistedMockSystemApiKeyState = { version: 1, keys }
  window.localStorage.setItem(MOCK_API_KEY_STORAGE_KEY, JSON.stringify(value))
}

const hotMockSystemApiKeys = import.meta.hot?.data.mockSystemApiKeys as
  | MockSystemApiKey[]
  | undefined
const mockApiKeyState: MockSystemApiKey[] =
  hotMockSystemApiKeys ?? readStoredKeys() ?? [...initialApiKeys]

if (import.meta.hot) {
  import.meta.hot.dispose((data) => {
    data.mockSystemApiKeys = mockApiKeyState
  })
}

export const mockSystemApiKeys = mockApiKeyState

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

function toApiKeyDTO(key: MockSystemApiKey, includeKey = false): ApiKeyDTO {
  return {
    id: key.id,
    name: key.name,
    key: includeKey ? key.key : maskApiKey(key.key),
    created_at: key.created_at,
    last_used_at: key.last_used_at,
  }
}

export function listMockSystemApiKeys(): ApiKeyDTO[] {
  return [...mockSystemApiKeys]
    .sort(
      (a, b) =>
        b.created_at.localeCompare(a.created_at) || Number(b.id) - Number(a.id)
    )
    .map((key) => toApiKeyDTO(key))
}

export function createMockSystemApiKey(name: string): ApiKeyDTO {
  const id = String(
    Math.max(0, ...mockSystemApiKeys.map((key) => Number(key.id))) + 1
  )
  const key: MockSystemApiKey = {
    id,
    name,
    key: mockApiKeyValue(id),
    created_at: new Date().toISOString(),
    last_used_at: null,
  }
  mockSystemApiKeys.push(key)
  persistKeys(mockSystemApiKeys)
  return toApiKeyDTO(key, true)
}

export function deleteMockSystemApiKey(id: string): boolean {
  const index = mockSystemApiKeys.findIndex((key) => key.id === id)
  if (index < 0) return false
  mockSystemApiKeys.splice(index, 1)
  persistKeys(mockSystemApiKeys)
  return true
}
