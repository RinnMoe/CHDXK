import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  getCurrentUser,
  login as loginApi,
  logout as logoutApi,
  register as registerApi,
  resetPassword as resetPasswordApi,
  sendRegisterCode as sendRegisterCodeApi,
  sendResetCode as sendResetCodeApi,
  type LoginCommand,
  type RegisterCommand,
  type ResetPasswordCommand,
  type SendRegisterCodeCommand,
  type SendResetCodeCommand,
} from "@/api/auth"

export function authMeQueryOptions() {
  return {
    queryKey: ["auth", "me"] as const,
    // The backend introspects the res-hub session on every request. Keep the
    // browser cache short-lived so role or account changes become visible.
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 0,
    refetchOnWindowFocus: true,
  }
}

export function useCurrentUser() {
  return useQuery(authMeQueryOptions())
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: LoginCommand) => loginApi(cmd),
    onSuccess: (user) => {
      qc.setQueryData(["auth", "me"], user)
    },
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: logoutApi,
    onSuccess: () => {
      qc.removeQueries({ queryKey: ["auth"] })
    },
  })
}

export function useRegister() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: RegisterCommand) => registerApi(cmd),
    onSuccess: (user) => {
      qc.setQueryData(["auth", "me"], user)
    },
  })
}

export function useSendRegisterCode() {
  return useMutation({
    mutationFn: (cmd: SendRegisterCodeCommand) => sendRegisterCodeApi(cmd),
  })
}

export function useSendResetCode() {
  return useMutation({
    mutationFn: (cmd: SendResetCodeCommand) => sendResetCodeApi(cmd),
  })
}

export function useResetPassword() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: ResetPasswordCommand) => resetPasswordApi(cmd),
    onSuccess: () => {
      qc.removeQueries({ queryKey: ["auth"] })
    },
  })
}
