import { normalizeAuthUser, type AuthUserDTO } from "@/api/auth"

interface MockUser {
  id: number
  username: string
  email: string
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
    email: "demo@chd.edu.cn",
    role: "user",
  }),
  normalizeMockUser({
    id: 2,
    username: "admin",
    email: "admin@chd.edu.cn",
    role: "super_admin",
  }),
]

interface MockAuthState {
  users: MockUser[]
  session: { userID: number | null }
}

interface PersistedMockAuthState {
  version: 1
  users: PersistedMockUser[]
  session: { userID: number | null }
}

const MOCK_AUTH_STORAGE_KEY = "jcourse:mock-auth-state"

function createMockAuthState(): MockAuthState {
  return {
    users: [...initialUsers],
    session: { userID: null },
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

export function findUserByID(id: number): MockUser | undefined {
  return mockUsers.find((u) => u.id === id)
}

export function setMockSessionUserID(userID: number | null) {
  mockSession.userID = userID
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
