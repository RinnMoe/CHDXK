import type { PaginatedResult } from "./types"
import type { CourseListItemDTO } from "./course"
import { apiClient } from "./client"

const BASE = "/api"

export const VoteLike = 1
export const VoteDislike = -1
export const VoteNeutral = 0

export type VoteType = typeof VoteLike | typeof VoteDislike | typeof VoteNeutral

export interface VoteStats {
  like_count: number
  dislike_count: number
  my_vote?: number
}

export interface ReviewDTO {
  id: number
  course?: CourseListItemDTO
  course_id: number
  user_id?: number
  semester?: string
  score: string
  rating: number
  content: string
  vote: VoteStats
  created_at: string
  updated_at: string
}

export interface ReviewRevisionDTO {
  id: number
  review_id: number
  course_id: number
  user_id: number
  semester: string
  score: string
  rating: number
  content: string
  created_at: string
}

export interface ReviewListFilter {
  semester?: string
  rating?: number
  order_by?: "like_count" | "created_at"
  ascend?: boolean
  page?: number
  page_size?: number
  [key: string]: unknown
}

export interface CreateReviewCommand {
  course_id: number
  semester?: string
  rating: number
  content: string
  score?: string
}

export interface UpdateReviewCommand {
  semester?: string
  rating: number
  content: string
  score?: string
}

function buildQuery(filter: Record<string, unknown>): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filter)) {
    if (value === undefined || value === null || value === "") continue
    if (Array.isArray(value)) {
      for (const v of value) params.append(key, String(v))
    } else if (typeof value === "boolean") {
      params.append(key, String(Number(value)))
    } else {
      params.append(key, String(value))
    }
  }
  const q = params.toString()
  return q ? `?${q}` : ""
}

export function listLatestReviews(
  filter: ReviewListFilter = {}
): Promise<PaginatedResult<ReviewDTO>> {
  return apiClient(`${BASE}/review/latest${buildQuery(filter)}`)
}

export function listFollowedReviews(
  filter: ReviewListFilter = {}
): Promise<PaginatedResult<ReviewDTO>> {
  return apiClient(`${BASE}/review/followed${buildQuery(filter)}`)
}

export function getReview(reviewID: number): Promise<ReviewDTO> {
  return apiClient(`${BASE}/review/${reviewID}`)
}

export function listReviewRevisions(
  reviewID: number
): Promise<ReviewRevisionDTO[]> {
  return apiClient(`${BASE}/review/${reviewID}/revisions`)
}

export function createReview(
  cmd: CreateReviewCommand
): Promise<{ message: string }> {
  return apiClient(`${BASE}/review/`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}

export function updateReview(
  reviewID: number,
  cmd: UpdateReviewCommand
): Promise<{ message: string }> {
  return apiClient(`${BASE}/review/${reviewID}`, {
    method: "PUT",
    body: JSON.stringify(cmd),
  })
}

export function deleteReview(reviewID: number): Promise<{ message: string }> {
  return apiClient(`${BASE}/review/${reviewID}`, {
    method: "DELETE",
  })
}

export function voteReview(
  reviewID: number,
  voteType: VoteType
): Promise<{ message: string }> {
  return apiClient(`${BASE}/review/${reviewID}/vote`, {
    method: "POST",
    body: JSON.stringify({ vote_type: voteType }),
  })
}

export function listUserReviews(
  userID: number,
  filter: ReviewListFilter = {}
): Promise<PaginatedResult<ReviewDTO>> {
  return apiClient(`${BASE}/user/${userID}/reviews${buildQuery(filter)}`)
}
