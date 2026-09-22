import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  createSystemApiKey,
  deleteSystemApiKey,
  listSystemApiKeys,
  type CreateApiKeyCommand,
} from "@/api/api-key"

export function useSystemApiKeys(enabled = true) {
  return useQuery({
    queryKey: ["api-keys", "system"],
    queryFn: listSystemApiKeys,
    enabled,
  })
}

export function useCreateSystemApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: CreateApiKeyCommand) => createSystemApiKey(cmd),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["api-keys", "system"] })
    },
  })
}

export function useDeleteSystemApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteSystemApiKey(id),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["api-keys", "system"] })
    },
  })
}
