import {
  keepPreviousData,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query"

import {
  createCanteenReview,
  deleteCanteenReview,
  listCanteenStoreReviews,
  listMyCanteenReviews,
  updateCanteenReview,
  voteCanteenReview,
  type CanteenReviewInput,
  type CanteenReviewListFilter,
  type CanteenReviewUpdateInput,
} from "@/api/canteen"
import { canteenQueryKeys } from "@/hooks/use-canteen-catalog"

const reviewKeys = {
  store: (storeID: number, filter: CanteenReviewListFilter) =>
    [...canteenQueryKeys.all, "reviews", storeID, filter] as const,
  mine: (page: number) => [...canteenQueryKeys.all, "my-reviews", page] as const,
}

export function useCanteenStoreReviews(
  storeID: number,
  filter: CanteenReviewListFilter = {}
) {
  return useInfiniteQuery({
    queryKey: reviewKeys.store(storeID, filter),
    queryFn: ({ pageParam }) =>
      listCanteenStoreReviews(storeID, {
        ...filter,
        cursor: pageParam || undefined,
      }),
    initialPageParam: "",
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
    enabled: storeID > 0,
  })
}

export function useMyCanteenReviews(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: reviewKeys.mine(page),
    queryFn: () => listMyCanteenReviews(page, pageSize),
    placeholderData: keepPreviousData,
  })
}

function useCanteenReviewInvalidate() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: canteenQueryKeys.all })
}

export function useCreateCanteenReview(storeID: number) {
  const invalidate = useCanteenReviewInvalidate()
  return useMutation({
    mutationFn: ({
      input,
      idempotencyKey,
    }: {
      input: CanteenReviewInput
      idempotencyKey: string
    }) => createCanteenReview(storeID, input, idempotencyKey),
    onSuccess: () => void invalidate(),
  })
}

export function useUpdateCanteenReview(reviewID: number) {
  const invalidate = useCanteenReviewInvalidate()
  return useMutation({
    mutationFn: (input: CanteenReviewUpdateInput) =>
      updateCanteenReview(reviewID, input),
    onSuccess: () => void invalidate(),
  })
}

export function useDeleteCanteenReview() {
  const invalidate = useCanteenReviewInvalidate()
  return useMutation({
    mutationFn: (reviewID: number) => deleteCanteenReview(reviewID),
    onSuccess: () => void invalidate(),
  })
}

export function useVoteCanteenReview() {
  const invalidate = useCanteenReviewInvalidate()
  return useMutation({
    mutationFn: ({
      reviewID,
      voteType,
    }: {
      reviewID: number
      voteType: number
    }) => voteCanteenReview(reviewID, voteType),
    onSuccess: () => void invalidate(),
  })
}
