import type { AuthUserDTO } from "@/api/auth"

const authSnapshotKey = "jcourse:last-auth-user"
const maxSnapshotAgeMs = 7 * 24 * 60 * 60 * 1000

type AuthSnapshot = {
  version: 1
  savedAt: number
  user: AuthUserDTO
}

function isAuthUser(value: unknown): value is AuthUserDTO {
  if (!value || typeof value !== "object") return false
  const user = value as Partial<AuthUserDTO>
  return (
    typeof user.id === "number" &&
    typeof user.username === "string" &&
    typeof user.email === "string" &&
    typeof user.role === "string"
  )
}

export function loadAuthSnapshot(): AuthUserDTO | null {
  try {
    const raw = window.localStorage.getItem(authSnapshotKey)
    if (!raw) return null

    const snapshot = JSON.parse(raw) as Partial<AuthSnapshot>
    if (snapshot.version !== 1 || !isAuthUser(snapshot.user)) return null
    if (typeof snapshot.savedAt !== "number") return null
    if (Date.now() - snapshot.savedAt > maxSnapshotAgeMs) {
      removeAuthSnapshot()
      return null
    }

    return snapshot.user
  } catch {
    return null
  }
}

export function saveAuthSnapshot(user: AuthUserDTO) {
  const snapshot: AuthSnapshot = {
    version: 1,
    savedAt: Date.now(),
    user,
  }

  try {
    window.localStorage.setItem(authSnapshotKey, JSON.stringify(snapshot))
  } catch {
    // ignore storage failures
  }
}

export function removeAuthSnapshot() {
  try {
    window.localStorage.removeItem(authSnapshotKey)
  } catch {
    // ignore storage failures
  }
}
