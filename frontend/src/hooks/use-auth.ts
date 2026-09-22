import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  getCurrentUser,
  logout as logoutApi,
} from "@/api/auth"

export function authMeQueryOptions() {
  return {
    queryKey: ["auth", "me"] as const,
    // The backend resolves the res-hub OIDC access token on every request. Keep the
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

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: logoutApi,
    onSuccess: () => {
      qc.removeQueries({ queryKey: ["auth"] })
    },
  })
}
