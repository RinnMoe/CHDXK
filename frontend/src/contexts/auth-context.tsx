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
  useLogin,
  useLogout,
  useRegister,
  useResetPassword,
  useSendRegisterCode,
  useSendResetCode,
} from "@/hooks/use-auth"
import type {
  AuthUserDTO,
  LoginCommand,
  RegisterCommand,
  ResetPasswordCommand,
} from "@/api/auth"
import { removeAuthSnapshot, saveAuthSnapshot } from "@/lib/auth-snapshot"
import { clearOfflineReadCaches } from "@/lib/offline-cache"

interface AuthContextValue {
  user: AuthUserDTO | null | undefined
  isLoading: boolean
  login: (cmd: LoginCommand) => Promise<AuthUserDTO>
  logout: () => Promise<void>
  register: (cmd: RegisterCommand) => Promise<AuthUserDTO>
  sendRegisterCode: (email: string) => Promise<void>
  sendResetCode: (email: string) => Promise<void>
  resetPassword: (cmd: ResetPasswordCommand) => Promise<void>
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
  const loginMutation = useLogin()
  const logoutMutation = useLogout()
  const registerMutation = useRegister()
  const sendRegisterCodeMutation = useSendRegisterCode()
  const sendResetCodeMutation = useSendResetCode()
  const resetPasswordMutation = useResetPassword()

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
    login: async (cmd) => {
      const loggedInUser = await loginMutation.mutateAsync(cmd)
      saveAuthSnapshot(loggedInUser)
      await clearBackendInterfaceCaches(queryClient, loggedInUser)
      return loggedInUser
    },
    logout: async () => {
      await logoutMutation.mutateAsync()
      removeAuthSnapshot()
      await clearBackendInterfaceCaches(queryClient, null)
    },
    register: async (cmd) => {
      const registeredUser = await registerMutation.mutateAsync(cmd)
      saveAuthSnapshot(registeredUser)
      await clearBackendInterfaceCaches(queryClient, registeredUser)
      return registeredUser
    },
    sendRegisterCode: async (email) => {
      await sendRegisterCodeMutation.mutateAsync({ email })
    },
    sendResetCode: async (email) => {
      await sendResetCodeMutation.mutateAsync({ email })
    },
    resetPassword: async (cmd) => {
      await resetPasswordMutation.mutateAsync(cmd)
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
