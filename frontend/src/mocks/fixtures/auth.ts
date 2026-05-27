import { normalizeAuthUser, type AuthUserDTO } from "@/api/auth"

interface MockUser {
  id: number
  username: string
  email: string
  password: string
  role: string
  is_admin: () => boolean
  is_super_admin: () => boolean
}

type PersistedMockUser = Omit<MockUser, "is_admin" | "is_super_admin">

function normalizeMockUser(user: PersistedMockUser): MockUser {
  return {
    ...user,
    is_admin: () => user.role === "admin" || user.role === "super_admin",
    is_super_admin: () => user.role === "super_admin",
  }
}

const initialUsers: MockUser[] = [
  normalizeMockUser({
    id: 1,
    username: "demo",
    email: "demo@sjtu.edu.cn",
    password: "password",
    role: "user",
  }),
  normalizeMockUser({
    id: 2,
    username: "admin",
    email: "admin@sjtu.edu.cn",
    password: "password",
    role: "super_admin",
  }),
]

interface MockAuthState {
  users: MockUser[]
  session: { userID: number | null }
  codes: Map<string, { code: string; sentAt: number }>
}

interface PersistedMockAuthState {
  version: 1
  users: PersistedMockUser[]
  session: { userID: number | null }
  codes: [string, { code: string; sentAt: number }][]
}

const MOCK_AUTH_STORAGE_KEY = "jcourse:mock-auth-state"

function createMockAuthState(): MockAuthState {
  return {
    users: [...initialUsers],
    session: { userID: null },
    codes: new Map<string, { code: string; sentAt: number }>(),
  }
}

function readStoredMockAuthState(): MockAuthState | null {
  if (typeof window === "undefined") return null

  const raw = window.localStorage.getItem(MOCK_AUTH_STORAGE_KEY)
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<PersistedMockAuthState>
    if (
      parsed.version !== 1 ||
      !Array.isArray(parsed.users) ||
      !parsed.session
    ) {
      return null
    }

    return {
      users: parsed.users.map(normalizeMockUser),
      session: { userID: parsed.session.userID ?? null },
      codes: new Map(parsed.codes ?? []),
    }
  } catch {
    window.localStorage.removeItem(MOCK_AUTH_STORAGE_KEY)
    return null
  }
}

function persistMockAuthState() {
  if (typeof window === "undefined") return

  const persisted: PersistedMockAuthState = {
    version: 1,
    users: mockAuthState.users,
    session: mockAuthState.session,
    codes: Array.from(mockAuthState.codes.entries()),
  }
  window.localStorage.setItem(MOCK_AUTH_STORAGE_KEY, JSON.stringify(persisted))
}

const mockAuthState: MockAuthState =
  import.meta.hot?.data.mockAuthState ??
  readStoredMockAuthState() ??
  createMockAuthState()

if (import.meta.hot) {
  import.meta.hot.dispose((data) => {
    data.mockAuthState = mockAuthState
  })
}

export const mockUsers = mockAuthState.users

export const mockSession = mockAuthState.session

export const mockCodes = mockAuthState.codes

const MOCK_CODE = "123456"

export function setMockCode(email: string) {
  mockCodes.set(email, { code: MOCK_CODE, sentAt: Date.now() })
  persistMockAuthState()
  return MOCK_CODE
}

export function consumeMockCode(email: string, code: string): boolean {
  const entry = mockCodes.get(email)
  if (!entry) return false
  if (entry.code !== code) return false
  mockCodes.delete(email)
  persistMockAuthState()
  return true
}

export function findUserByEmail(email: string): MockUser | undefined {
  return mockUsers.find((u) => u.email === email)
}

export function findUserByID(id: number): MockUser | undefined {
  return mockUsers.find((u) => u.id === id)
}

export function addUser(email: string, password: string): MockUser {
  const id = Math.max(0, ...mockUsers.map((u) => u.id)) + 1
  const username = email.split("@")[0]
  const user = normalizeMockUser({ id, username, email, password, role: "user" })
  mockUsers.push(user)
  persistMockAuthState()
  return user
}

export function setMockSessionUserID(userID: number | null) {
  mockSession.userID = userID
  persistMockAuthState()
}

export function setMockUserPassword(user: MockUser, password: string) {
  user.password = password
  persistMockAuthState()
}

export function toAuthUserDTO(u: MockUser): AuthUserDTO {
  return normalizeAuthUser({
    id: u.id,
    username: u.username,
    email: u.email,
    role: u.role,
  })
}
