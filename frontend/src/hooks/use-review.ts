import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  createReview,
  deleteReview,
  getReview,
  listFollowedReviews,
  listLatestReviews,
  listUserReviews,
  updateReview,
  voteReview,
  type CreateReviewCommand,
  type ReviewListFilter,
  type UpdateReviewCommand,
  type VoteType,
} from "@/api/review"

export function useLatestReviews(filter: ReviewListFilter = {}) {
  return useQuery({
    queryKey: ["reviews", "latest", filter],
    queryFn: () => listLatestReviews(filter),
  })
}

export function useFollowedReviews(filter: ReviewListFilter = {}) {
  return useQuery({
    queryKey: ["reviews", "followed", filter],
    queryFn: () => listFollowedReviews(filter),
  })
}

export function useUserReviews(userID: number, filter: ReviewListFilter = {}) {
  return useQuery({
    queryKey: ["reviews", "user", userID, filter],
    queryFn: () => listUserReviews(userID, filter),
    enabled: !!userID,
  })
}

export function useReview(reviewID: number) {
  return useQuery({
    queryKey: ["review", reviewID],
    queryFn: () => getReview(reviewID),
    enabled: !!reviewID,
  })
}

export function useCreateReview() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (cmd: CreateReviewCommand) => createReview(cmd),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reviews"] })
      queryClient.invalidateQueries({ queryKey: ["courses"] })
    },
  })
}

export function useUpdateReview() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      reviewID,
      cmd,
    }: {
      reviewID: number
      cmd: UpdateReviewCommand
    }) => updateReview(reviewID, cmd),
    onSuccess: (_, { reviewID }) => {
      queryClient.invalidateQueries({ queryKey: ["reviews"] })
      queryClient.invalidateQueries({ queryKey: ["review", reviewID] })
    },
  })
}

export function useDeleteReview() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (reviewID: number) => deleteReview(reviewID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reviews"] })
      queryClient.invalidateQueries({ queryKey: ["courses"] })
    },
  })
}

export function useVoteReview() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      reviewID,
      voteType,
    }: {
      reviewID: number
      voteType: VoteType
    }) => voteReview(reviewID, voteType),
    onMutate: async ({ reviewID, voteType }) => {
      await queryClient.cancelQueries({ queryKey: ["review", reviewID] })
      const prev = queryClient.getQueryData(["review", reviewID])
      queryClient.setQueryData(["review", reviewID], (old: unknown) => {
        if (!old || typeof old !== "object") return old
        const o = old as { vote?: { like_count: number; dislike_count: number; my_vote?: number } }
        return {
          ...o,
          vote: {
            ...o.vote,
            my_vote: voteType,
          },
        }
      })
      return { prev }
    },
    onError: (_err, { reviewID }, context) => {
      if (context?.prev) {
        queryClient.setQueryData(["review", reviewID], context.prev)
      }
    },
    onSettled: (_data, _err, { reviewID }) => {
      queryClient.invalidateQueries({ queryKey: ["review", reviewID] })
    },
  })
}
