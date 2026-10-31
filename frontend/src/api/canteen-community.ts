import type {
  CanteenAdminMediaDTO,
  CanteenContributionsDTO,
  CanteenDishDTO,
  CanteenDishReviewDTO,
  CanteenMediaDTO,
  CanteenMediaAssetDTO,
  CanteenMediaUploadDTO,
  CanteenReportDTO,
} from "./canteen"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"
import { buildQuery } from "./query"

export interface CanteenDishDetailDTO {
  dish: CanteenDishDTO
  photos: CanteenMediaDTO[]
  reviews: CanteenDishReviewDTO[]
  review_total: number
}

export interface CanteenDishReviewInput {
  review_event_id?: number
  taste: number
  portion: number
  value: number
  content: string
  media_ids?: number[]
}

export function getStoreDishDetail(storeID: number, dishID: number) {
  return apiClient<CanteenDishDetailDTO>(
    `${BASE_URL}/canteen/stores/${storeID}/dishes/${dishID}`
  )
}


export interface CanteenMediaIntentInput {
  purpose: "review" | "dish" | "store_cover"
  mime: string
  bytes: number
}

export function listStoreDishes(storeID: number) {
  return apiClient<{ items: CanteenDishDTO[] }>(
    `${BASE_URL}/canteen/stores/${storeID}/dishes`
  )
}

export function suggestCanteenDishes(storeID: number, q: string) {
  return apiClient<{ items: CanteenDishDTO[] }>(
    `${BASE_URL}/canteen/dishes/suggest${buildQuery({ store_id: storeID, q })}`
  )
}

export function createCanteenDish(storeID: number, name: string) {
  return apiClient<{ id: number; name: string; status: string; existing: boolean }>(
    `${BASE_URL}/canteen/stores/${storeID}/dishes`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }
  )
}

export function createCanteenDishReview(
  dishID: number,
  input: CanteenDishReviewInput
) {
  return apiClient<{ id: number; dish_id: number; overall_score: number }>(
    `${BASE_URL}/canteen/dishes/${dishID}/reviews`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }
  )
}

export function createCanteenMediaIntent(input: CanteenMediaIntentInput) {
  return apiClient<CanteenMediaUploadDTO>(
    `${BASE_URL}/canteen/media/upload-intents`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }
  )
}

export function putCanteenMediaContent(mediaID: number, file: Blob) {
  return apiClient<CanteenMediaAssetDTO>(`${BASE_URL}/canteen/media/${mediaID}/content`, {
    method: "PUT",
    headers: { "Content-Type": file.type || "application/octet-stream" },
    body: file,
  })
}

export function completeCanteenMedia(mediaID: number) {
  return apiClient<CanteenMediaAssetDTO>(
    `${BASE_URL}/canteen/media/${mediaID}/complete`,
    { method: "POST" }
  )
}

export function canteenMediaURL(mediaID: number) {
  return `${BASE_URL}/canteen/media/${mediaID}`
}

export function listMyCanteenContributions() {
  return apiClient<CanteenContributionsDTO>(
    `${BASE_URL}/canteen/contributions`
  )
}

export function createCanteenReport(input: {
  subject_type: "store" | "review" | "dish_review" | "media"
  subject_id: number
  reason_code: string
  detail: string
}) {
  return apiClient<{ id: number; status: string }>(`${BASE_URL}/canteen/reports`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
}

export function listAdminCanteenDishes(filter: {
  store_id?: number
  status?: string
  q?: string
}) {
  return apiClient<{ items: CanteenDishDTO[] }>(
    `${BASE_URL}/admin/canteen/dishes${buildQuery(filter as Record<string, unknown>)}`
  )
}

export function updateAdminCanteenDishStatus(dishID: number, status: string) {
  return apiClient<{ id: number; status: string }>(
    `${BASE_URL}/admin/canteen/dishes/${dishID}/status`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    }
  )
}

export function mergeAdminCanteenDish(dishID: number, targetID: number) {
  return apiClient<{ message: string }>(
    `${BASE_URL}/admin/canteen/dishes/${dishID}/merge`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ target_id: targetID }),
    }
  )
}

export function listAdminCanteenMedia(status = "", page = 1, pageSize = 20) {
  return apiClient<{
    items: CanteenAdminMediaDTO[]
    total: number
    page: number
    page_size: number
  }>(
    `${BASE_URL}/admin/canteen/media${buildQuery({
      status,
      page,
      page_size: pageSize,
    })}`
  )
}

export function reviewAdminCanteenMedia(
  mediaID: number,
  status: "approved" | "rejected",
  reason = ""
) {
  return apiClient<{ message: string }>(
    `${BASE_URL}/admin/canteen/media/${mediaID}/review`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status, reason }),
    }
  )
}

export function listAdminCanteenReports(status = "", page = 1, pageSize = 20) {
  return apiClient<{
    items: CanteenReportDTO[]
    total: number
    page: number
    page_size: number
  }>(
    `${BASE_URL}/admin/canteen/reports${buildQuery({
      status,
      page,
      page_size: pageSize,
    })}`
  )
}

export function resolveAdminCanteenReport(
  reportID: number,
  status: "resolved" | "dismissed",
  note = ""
) {
  return apiClient<{ message: string }>(
    `${BASE_URL}/admin/canteen/reports/${reportID}/resolve`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status, note }),
    }
  )
}
