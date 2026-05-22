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

export function useCurrentUser() {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 1000 * 60 * 5,
  })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: LoginCommand) => loginApi(cmd),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["auth"] })
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
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["auth"] })
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
