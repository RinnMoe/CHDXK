import type {
  CanteenDishDTO,
  CanteenHallDTO,
  CanteenReviewDTO,
  CanteenStoreListItemDTO,
} from "./canteen"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"
import { buildQuery } from "./query"

export interface CanteenHomeStatsDTO {
  store_count: number
  review_count: number
  dish_count: number
}

export interface CanteenHomeHeroDTO {
  type: string
  store_id: number
  store_name: string
  hall_name: string
  dish_id: number
  dish_name: string
  image_url: string
  reason: string
  rating_score: number
  rating_count: number
}

export interface CanteenHomeFeedDTO {
  type: string
  store_id: number
  store_name: string
  hall_name: string
  dish_id: number
  dish_name: string
  reason: string
  image_url: string
  rating_score: number
  rating_count: number
  tags: string[]
}

export interface CanteenHomeDTO {
  halls: CanteenHallDTO[]
  top_stores: CanteenStoreListItemDTO[]
  new_stores: CanteenStoreListItemDTO[]
  latest_reviews: CanteenReviewDTO[]
  hot_dishes: CanteenDishDTO[]
  hero: CanteenHomeHeroDTO | null
  feeds: CanteenHomeFeedDTO[]
  stats: CanteenHomeStatsDTO
}

export interface CanteenSearchDTO {
  stores: CanteenStoreListItemDTO[]
  dishes: CanteenDishDTO[]
}

export function getCanteenHome() {
  return apiClient<CanteenHomeDTO>(`${BASE_URL}/canteen/home`)
}

export function searchCanteen(q: string) {
  return apiClient<CanteenSearchDTO>(
    `${BASE_URL}/canteen/search${buildQuery({ q })}`
  )
}
