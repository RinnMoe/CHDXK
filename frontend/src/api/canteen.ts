import type { PaginatedResult } from "./types"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"
import { buildQuery } from "./query"

export interface CanteenHallDTO {
  id: number
  campus_key: string
  name: string
  short_name: string
  address: string
  status: string
  display_order: number
  store_count: number
}

export interface CanteenRatingDTO {
  count: number
  avg: number
  score: number
}

export interface CanteenDimensionsDTO {
  taste: number
  value: number
  queue: number
  hygiene: number
  service: number
}

export interface CanteenMySummaryDTO {
  weighted_score: number
  window_size: number
  lifetime_event_count: number
  last_visited_at: string | null
}

export interface CanteenMediaDTO {
  id: number
  url: string
  width: number
  height: number
}

export interface CanteenReviewDTO {
  id: number
  store_id: number
  store_name: string
  user_id: number
  visited_at: string
  taste: number
  value: number
  queue: number
  hygiene: number
  service: number
  overall_score: number
  content: string
  tags: string[]
  status: string
  like_count: number
  dislike_count: number
  my_vote: number
  media: CanteenMediaDTO[]
  created_at: string
  updated_at: string
}

export interface CanteenDishDTO {
  id: number
  store_id: number
  store_name: string
  name: string
  status: string
  rating_count: number
  rating_avg: number
  rating_score: number
  taste_avg: number
  portion_avg: number
  value_avg: number
  review_count: number
  cover_url: string
  photo_count: number
}

export interface CanteenStoreListItemDTO {
  id: number
  hall_id: number
  hall_name: string
  campus_key: string
  name: string
  floor: string
  categories: string[]
  status: string
  cover_url: string
  rating: CanteenRatingDTO
  review_count: number
  last_review_at: string | null
  confidence?: string
  summary_tags: string[]
  dish_count: number
  dish_photo_count: number
}

export interface CanteenStoreDetailDTO extends CanteenStoreListItemDTO {
  dimensions: CanteenDimensionsDTO
  dishes: CanteenDishDTO[]
  my_summary: CanteenMySummaryDTO | null
  my_latest_review: CanteenReviewDTO | null
}

export type CanteenStorePage = PaginatedResult<CanteenStoreListItemDTO>

export interface CanteenReviewPage {
  items: CanteenReviewDTO[]
  total: number
  page: number
  page_size: number
  next_cursor?: string
}

export interface CanteenReviewInput {
  visited_at?: string
  taste: number
  value: number
  queue: number
  hygiene: number
  service: number
  content: string
  tags: string[]
  media_ids?: number[]
}

export interface CanteenReviewUpdateInput extends CanteenReviewInput {
  expected_updated_at?: string
}

export interface CanteenReviewListFilter {
  order?: "latest" | "best" | "worst"
  filter?: "all" | "with_media" | "high" | "low"
  history?: boolean
  cursor?: string
  limit?: number
  with_media?: boolean
  page?: number
  page_size?: number
}

export interface CanteenStoreListFilter {
  q?: string
  campus?: string
  hall_id?: number
  category?: string
  sort?: "rating_score" | "rating_count" | "latest" | "name"
  page?: number
  page_size?: number
}

export interface CanteenHallInput {
  campus_key: string
  name: string
  short_name?: string
  address?: string
  status?: string
  display_order?: number
}

export interface CanteenStoreInput {
  hall_id: number
  name: string
  floor?: string
  categories?: string[]
  status?: string
}

export async function listCanteenHalls() {
  const result = await apiClient<{ items: CanteenHallDTO[] }>(
    `${BASE_URL}/canteen/halls`
  )
  return result.items
}

export function listCanteenStores(filter: CanteenStoreListFilter = {}) {
  return apiClient<CanteenStorePage>(
    `${BASE_URL}/canteen/stores${buildQuery(filter as Record<string, unknown>)}`
  )
}

export function listCanteenRankings(filter: CanteenStoreListFilter = {}) {
  return apiClient<CanteenStorePage>(
    `${BASE_URL}/canteen/rankings${buildQuery(filter as Record<string, unknown>)}`
  )
}

export function getCanteenStore(storeID: number) {
  return apiClient<CanteenStoreDetailDTO>(
    `${BASE_URL}/canteen/stores/${storeID}`
  )
}

export function createCanteenHall(input: CanteenHallInput) {
  return apiClient<{ id: number }>(`${BASE_URL}/admin/canteen/halls`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
}

export function updateCanteenHall(hallID: number, input: CanteenHallInput) {
  return apiClient<{ id: number }>(`${BASE_URL}/admin/canteen/halls/${hallID}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
}

export function listAdminCanteenStores(
  filter: CanteenStoreListFilter & { status?: string } = {}
) {
  return apiClient<CanteenStorePage>(
    `${BASE_URL}/admin/canteen/stores${buildQuery(filter as Record<string, unknown>)}`
  )
}

export function createCanteenStore(input: CanteenStoreInput) {
  return apiClient<{ id: number }>(`${BASE_URL}/admin/canteen/stores`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
}

export function updateCanteenStore(storeID: number, input: CanteenStoreInput) {
  return apiClient<{ id: number }>(
    `${BASE_URL}/admin/canteen/stores/${storeID}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }
  )
}

export function updateCanteenStoreStatus(storeID: number, status: string) {
  return apiClient<{ id: number; status: string }>(
    `${BASE_URL}/admin/canteen/stores/${storeID}/status`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    }
  )
}

export function mergeCanteenStore(storeID: number, targetID: number) {
  return apiClient<{ id: number; status: string }>(
    `${BASE_URL}/admin/canteen/stores/${storeID}/merge`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ target_id: targetID }),
    }
  )
}

export function listCanteenStoreReviews(
  storeID: number,
  filter: CanteenReviewListFilter = {}
) {
  return apiClient<CanteenReviewPage>(
    `${BASE_URL}/canteen/stores/${storeID}/reviews${buildQuery(
      filter as Record<string, unknown>
    )}`
  )
}

export function listMyCanteenReviews(page = 1, pageSize = 20) {
  return apiClient<CanteenReviewPage>(
    `${BASE_URL}/canteen/reviews/mine${buildQuery({ page, page_size: pageSize })}`
  )
}

export function createCanteenReview(
  storeID: number,
  input: CanteenReviewInput,
  idempotencyKey: string
) {
  return apiClient<{ created: boolean; review: CanteenReviewDTO }>(
    `${BASE_URL}/canteen/stores/${storeID}/reviews`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": idempotencyKey,
      },
      body: JSON.stringify(input),
    }
  )
}

export function updateCanteenReview(
  reviewID: number,
  input: CanteenReviewUpdateInput
) {
  return apiClient<{ review: CanteenReviewDTO }>(
    `${BASE_URL}/canteen/reviews/${reviewID}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }
  )
}

export function deleteCanteenReview(reviewID: number) {
  return apiClient<{ message: string }>(
    `${BASE_URL}/canteen/reviews/${reviewID}`,
    { method: "DELETE" }
  )
}

export function voteCanteenReview(reviewID: number, voteType: number) {
  return apiClient<{ message: string }>(
    `${BASE_URL}/canteen/reviews/${reviewID}/vote`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ vote_type: voteType }),
    }
  )
}


export interface CanteenDishReviewDTO {
  id: number
  dish_id: number
  dish_name: string
  store_id: number
  store_name: string
  taste: number
  portion: number
  value: number
  overall_score: number
  content: string
  status: string
  created_at: string
}


export interface CanteenMediaUploadDTO {
  media_id: number
  expires_at: string
  upload: {
    mode: string
    method: string
    url: string
    headers: Record<string, string>
  }
}

export interface CanteenMediaAssetDTO {
  id: number
  url: string
  mime: string
  byte_size: number
  width: number
  height: number
  status: string
  created_at: string
}

export interface CanteenAdminMediaDTO extends CanteenMediaAssetDTO {
  owner_user_id: number
}

export interface CanteenReportDTO {
  id: number
  reporter_user_id: number
  subject_type: string
  subject_id: number
  reason_code: string
  detail: string
  status: string
  handled_by: number | null
  handled_at: string | null
  resolution_note: string
  created_at: string
}

export interface CanteenContributionsDTO {
  dishes: CanteenDishReviewDTO[]
  media: CanteenMediaAssetDTO[]
  dish_total: number
  media_total: number
  page: number
  page_size: number
}

