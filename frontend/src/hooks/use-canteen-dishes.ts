import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import type { CanteenDishReviewInput } from "@/api/canteen-community"
import {
  createCanteenDish,
  createCanteenDishReview,
  getStoreDishDetail,
  listAdminCanteenDishes,
  listMyCanteenContributions,
  listStoreDishes,
  mergeAdminCanteenDish,
  updateAdminCanteenDishStatus,
} from "@/api/canteen-community"
import { canteenQueryKeys } from "@/hooks/use-canteen-catalog"

export function useStoreDishes(storeID: number) {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "dishes", storeID],
    queryFn: () => listStoreDishes(storeID),
    enabled: storeID > 0,
  })
}

export function useStoreDishDetail(storeID: number, dishID: number) {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "dish", storeID, dishID],
    queryFn: () => getStoreDishDetail(storeID, dishID),
    enabled: storeID > 0 && dishID > 0,
  })
}

export function useMyCanteenContributions() {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "contributions"],
    queryFn: listMyCanteenContributions,
  })
}

export function useAdminCanteenDishes(filter: {
  store_id?: number
  status?: string
  q?: string
}) {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "admin-dishes", filter],
    queryFn: () => listAdminCanteenDishes(filter),
    enabled: !!filter.store_id,
  })
}

function useCanteenInvalidate() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: canteenQueryKeys.all })
}

export function useCreateCanteenDish(storeID: number) {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: (name: string) => createCanteenDish(storeID, name),
    onSuccess: () => void invalidate(),
  })
}

export function useCreateCanteenDishReview(dishID: number) {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: (input: CanteenDishReviewInput) =>
      createCanteenDishReview(dishID, input),
    onSuccess: () => void invalidate(),
  })
}

export function useUpdateAdminCanteenDishStatus() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: ({ dishID, status }: { dishID: number; status: string }) =>
      updateAdminCanteenDishStatus(dishID, status),
    onSuccess: () => void invalidate(),
  })
}

export function useMergeAdminCanteenDish() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: ({ dishID, targetID }: { dishID: number; targetID: number }) =>
      mergeAdminCanteenDish(dishID, targetID),
    onSuccess: () => void invalidate(),
  })
}
