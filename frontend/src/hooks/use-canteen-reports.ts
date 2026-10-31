import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  createCanteenReport,
  listAdminCanteenReports,
  resolveAdminCanteenReport,
} from "@/api/canteen-community"
import { canteenQueryKeys } from "@/hooks/use-canteen-catalog"

export function useCreateCanteenReport() {
  return useMutation({
    mutationFn: (input: {
      subject_type: "store" | "review" | "dish_review" | "media"
      subject_id: number
      reason_code: string
      detail: string
    }) => createCanteenReport(input),
  })
}

export function useAdminCanteenReports(status = "", page = 1, pageSize = 20) {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "admin-reports", status, page, pageSize],
    queryFn: () => listAdminCanteenReports(status, page, pageSize),
  })
}

export function useResolveAdminCanteenReport() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      reportID,
      status,
      note,
    }: {
      reportID: number
      status: "resolved" | "dismissed"
      note?: string
    }) => resolveAdminCanteenReport(reportID, status, note),
    onSuccess: () =>
      void queryClient.invalidateQueries({ queryKey: canteenQueryKeys.all }),
  })
}
