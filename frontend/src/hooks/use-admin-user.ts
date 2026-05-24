import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  clearAdminUserSuspension,
  getAdminUserByEmail,
  suspendAdminUser,
  type SuspendAdminUserCommand,
} from "@/api/admin-user"

type SuspendAdminUserVariables = {
  userID: number
  cmd?: SuspendAdminUserCommand
}

export function useAdminUserByEmail(email: string) {
  const normalized = email.trim().toLowerCase()

  return useQuery({
    queryKey: ["admin-user", "by-email", normalized],
    queryFn: () => getAdminUserByEmail(normalized),
    enabled: normalized.length > 0,
    retry: false,
  })
}

export function useSuspendAdminUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ userID, cmd }: SuspendAdminUserVariables) =>
      suspendAdminUser(userID, cmd),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}

export function useClearAdminUserSuspension() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: clearAdminUserSuspension,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-user"] })
    },
  })
}
