import {
  createContext,
  useContext,
  useEffect,
  useRef,
  type ReactNode,
} from "react"
import { useQueryClient, type QueryClient } from "@tanstack/react-query"
import {
  useCurrentUser,
  useLogout,
} from "@/hooks/use-auth"
import type { AuthUserDTO } from "@/api/auth"
import { clearOfflineReadCaches } from "@/lib/offline-cache"

interface AuthContextValue {
  user: AuthUserDTO | null | undefined
  isLoading: boolean
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

async function clearBackendInterfaceCaches(
  queryClient: QueryClient,
  currentUser: AuthUserDTO | null
) {
  queryClient.removeQueries({
    predicate: (query) => query.queryKey[0] !== "auth",
  })
  queryClient.setQueryData(["auth", "me"], currentUser)
  await clearOfflineReadCaches()
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const { data: user, isLoading } = useCurrentUser()
  const previousUserID = useRef<number | null | undefined>(undefined)
  const logoutMutation = useLogout()

  useEffect(() => {
    void clearOfflineReadCaches()
  }, [])

  useEffect(() => {
    if (isLoading) return

    const currentUserID = user?.id ?? null
    if (previousUserID.current === undefined) {
      previousUserID.current = currentUserID
      return
    }

    if (previousUserID.current !== currentUserID) {
      previousUserID.current = currentUserID
      void clearBackendInterfaceCaches(queryClient, user ?? null)
    }
  }, [isLoading, queryClient, user])

  const value: AuthContextValue = {
    user,
    isLoading,
    logout: async () => {
      await logoutMutation.mutateAsync()
      await clearBackendInterfaceCaches(queryClient, null)
    },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return ctx
}
