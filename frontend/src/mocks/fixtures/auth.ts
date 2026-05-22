import type { AuthUserDTO } from "@/api/auth"

interface MockUser {
  id: number
  username: string
  email: string
  password: string
  role: string
}

const initialUsers: MockUser[] = [
  {
    id: 1,
    username: "demo",
    email: "demo@sjtu.edu.cn",
    password: "password",
    role: "user",
  },
  {
    id: 2,
    username: "admin",
    email: "admin@sjtu.edu.cn",
    password: "password",
    role: "admin",
  },
]

export const mockUsers: MockUser[] = [...initialUsers]

export const mockSession: { userID: number | null } = { userID: null }

export const mockCodes = new Map<string, { code: string; sentAt: number }>()

const MOCK_CODE = "123456"

export function setMockCode(email: string) {
  mockCodes.set(email, { code: MOCK_CODE, sentAt: Date.now() })
  return MOCK_CODE
}

export function consumeMockCode(email: string, code: string): boolean {
  const entry = mockCodes.get(email)
  if (!entry) return false
  if (entry.code !== code) return false
  mockCodes.delete(email)
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
  const user: MockUser = { id, username, email, password, role: "user" }
  mockUsers.push(user)
  return user
}

export function toAuthUserDTO(u: MockUser): AuthUserDTO {
  return { id: u.id, username: u.username, email: u.email, role: u.role }
}
