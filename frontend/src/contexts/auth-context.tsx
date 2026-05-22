import { createContext, useContext, type ReactNode } from "react"
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

export function AuthProvider({ children }: { children: ReactNode }) {
  const { data: user, isLoading } = useCurrentUser()
  const loginMutation = useLogin()
  const logoutMutation = useLogout()
  const registerMutation = useRegister()
  const sendRegisterCodeMutation = useSendRegisterCode()
  const sendResetCodeMutation = useSendResetCode()
  const resetPasswordMutation = useResetPassword()

  const value: AuthContextValue = {
    user,
    isLoading,
    login: (cmd) => loginMutation.mutateAsync(cmd),
    logout: async () => {
      await logoutMutation.mutateAsync()
    },
    register: (cmd) => registerMutation.mutateAsync(cmd),
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
