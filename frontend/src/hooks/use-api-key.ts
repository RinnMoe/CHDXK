import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  createApiKey,
  deleteApiKey,
  listApiKeys,
  type CreateApiKeyCommand,
} from "@/api/api-key"

export function useApiKeys(enabled = true) {
  return useQuery({
    queryKey: ["api-keys"],
    queryFn: listApiKeys,
    enabled,
  })
}

export function useCreateApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: CreateApiKeyCommand) => createApiKey(cmd),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["api-keys"] })
    },
  })
}

export function useDeleteApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => deleteApiKey(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["api-keys"] })
    },
  })
}
