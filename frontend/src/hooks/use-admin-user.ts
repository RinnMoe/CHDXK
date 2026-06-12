import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  clearAdminUserSuspension,
  getAdminUser,
  grantAdminUser,
  listAdminUsers,
  resetAdminUserPassword,
  revokeAdminUser,
  suspendAdminUser,
  type AdminUserLookup,
  type ResetAdminUserPasswordCommand,
  type SuspendAdminUserCommand,
} from "@/api/admin-user"

type SuspendAdminUserVariables = {
  userID: number
  cmd: SuspendAdminUserCommand
}

type ResetAdminUserPasswordVariables = {
  userID: number
  cmd: ResetAdminUserPasswordCommand
}

export function useAdminUser(lookup: AdminUserLookup) {
  const normalized: AdminUserLookup = {
    email: lookup.email?.trim().toLowerCase() || undefined,
    username: lookup.username?.trim() || undefined,
    review_id: lookup.review_id,
  }
  const enabled = Boolean(
    normalized.email || normalized.username || normalized.review_id
  )

  return useQuery({
    queryKey: ["admin-user", "query", normalized],
    queryFn: () => getAdminUser(normalized),
    enabled,
    retry: false,
  })
}

export function useAdminUsers() {
  return useQuery({
    queryKey: ["admin-user", "admins"],
    queryFn: listAdminUsers,
  })
}

export function useSuspendAdminUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ userID, cmd }: SuspendAdminUserVariables) =>
      suspendAdminUser(userID, cmd),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}

export function useClearAdminUserSuspension() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: clearAdminUserSuspension,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}

export function useGrantAdminUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: grantAdminUser,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}

export function useRevokeAdminUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: revokeAdminUser,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}

export function useResetAdminUserPassword() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ userID, cmd }: ResetAdminUserPasswordVariables) =>
      resetAdminUserPassword(userID, cmd),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}
