import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  completeCanteenMedia,
  createCanteenMediaIntent,
  listAdminCanteenMedia,
  putCanteenMediaContent,
  reviewAdminCanteenMedia,
} from "@/api/canteen-community"
import { canteenQueryKeys } from "@/hooks/use-canteen-catalog"

export function useUploadCanteenMedia(
  purpose: "review" | "dish" = "review"
) {
  return useMutation({
    mutationFn: async (file: File) => {
      const intent = await createCanteenMediaIntent({
        purpose,
        mime: file.type,
        bytes: file.size,
      })
      const asset = await putCanteenMediaContent(intent.media_id, file)
      await completeCanteenMedia(intent.media_id)
      return asset
    },
  })
}

export function useAdminCanteenMedia(status = "", page = 1, pageSize = 20) {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "admin-media", status, page, pageSize],
    queryFn: () => listAdminCanteenMedia(status, page, pageSize),
  })
}

export function useReviewAdminCanteenMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      mediaID,
      status,
      reason,
    }: {
      mediaID: number
      status: "approved" | "rejected"
      reason?: string
    }) => reviewAdminCanteenMedia(mediaID, status, reason),
    onSuccess: () =>
      void queryClient.invalidateQueries({ queryKey: canteenQueryKeys.all }),
  })
}
