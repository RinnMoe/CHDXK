import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  createTransfer,
  getUserPoints,
  previewTransfer,
  type CreatePointTransferCommand,
  type PointRecordListFilter,
  type PreviewTransferParams,
} from "@/api/point"

export function useUserPoints(userID: number, filter: PointRecordListFilter = {}) {
  return useQuery({
    queryKey: ["points", userID, filter],
    queryFn: () => getUserPoints(userID, filter),
    enabled: !!userID,
  })
}

export function usePreviewTransfer() {
  return useMutation({
    mutationFn: (params: PreviewTransferParams) => previewTransfer(params),
  })
}

export function useCreateTransfer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (cmd: CreatePointTransferCommand) => createTransfer(cmd),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["points"] })
    },
  })
}
