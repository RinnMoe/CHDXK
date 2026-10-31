import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  createCanteenHall,
  createCanteenStore,
  getCanteenStore,
  listAdminCanteenStores,
  listCanteenHalls,
  listCanteenRankings,
  listCanteenStores,
  mergeCanteenStore,
  updateCanteenHall,
  updateCanteenStore,
  updateCanteenStoreStatus,
  type CanteenHallInput,
  type CanteenStoreInput,
  type CanteenStoreListFilter,
} from "@/api/canteen"
import { getCanteenHome, searchCanteen } from "@/api/canteen-home"

export const canteenQueryKeys = {
  all: ["canteen"] as const,
  home: () => [...canteenQueryKeys.all, "home"] as const,
  halls: () => [...canteenQueryKeys.all, "halls"] as const,
  stores: (filter: CanteenStoreListFilter) =>
    [...canteenQueryKeys.all, "stores", filter] as const,
  rankings: (filter: CanteenStoreListFilter) =>
    [...canteenQueryKeys.all, "rankings", filter] as const,
  store: (storeID: number) =>
    [...canteenQueryKeys.all, "store", storeID] as const,
  adminStores: (filter: CanteenStoreListFilter) =>
    [...canteenQueryKeys.all, "admin-stores", filter] as const,
}

export function useCanteenHome() {
  return useQuery({ queryKey: canteenQueryKeys.home(), queryFn: getCanteenHome })
}

export function useCanteenHalls() {
  return useQuery({ queryKey: canteenQueryKeys.halls(), queryFn: listCanteenHalls })
}

export function useCanteenStores(filter: CanteenStoreListFilter = {}) {
  return useQuery({
    queryKey: canteenQueryKeys.stores(filter),
    queryFn: () => listCanteenStores(filter),
    placeholderData: keepPreviousData,
  })
}

export function useCanteenRankings(filter: CanteenStoreListFilter = {}) {
  return useQuery({
    queryKey: canteenQueryKeys.rankings(filter),
    queryFn: () => listCanteenRankings(filter),
    placeholderData: keepPreviousData,
  })
}

export function useCanteenStore(storeID: number) {
  return useQuery({
    queryKey: canteenQueryKeys.store(storeID),
    queryFn: () => getCanteenStore(storeID),
    enabled: storeID > 0,
  })
}

export function useCanteenSearch(query: string) {
  return useQuery({
    queryKey: [...canteenQueryKeys.all, "search", query],
    queryFn: () => searchCanteen(query),
    enabled: query.trim().length > 0,
  })
}

export function useAdminCanteenStores(
  filter: CanteenStoreListFilter & { status?: string } = {}
) {
  return useQuery({
    queryKey: canteenQueryKeys.adminStores(filter),
    queryFn: () => listAdminCanteenStores(filter),
    placeholderData: keepPreviousData,
  })
}

function useCanteenInvalidate() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: canteenQueryKeys.all })
}

export function useCreateCanteenHall() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: (input: CanteenHallInput) => createCanteenHall(input),
    onSuccess: () => void invalidate(),
  })
}

export function useUpdateCanteenHall() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: ({ hallID, input }: { hallID: number; input: CanteenHallInput }) =>
      updateCanteenHall(hallID, input),
    onSuccess: () => void invalidate(),
  })
}

export function useCreateCanteenStore() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: (input: CanteenStoreInput) => createCanteenStore(input),
    onSuccess: () => void invalidate(),
  })
}

export function useUpdateCanteenStore() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: ({ storeID, input }: { storeID: number; input: CanteenStoreInput }) =>
      updateCanteenStore(storeID, input),
    onSuccess: () => void invalidate(),
  })
}

export function useUpdateCanteenStoreStatus() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: ({ storeID, status }: { storeID: number; status: string }) =>
      updateCanteenStoreStatus(storeID, status),
    onSuccess: () => void invalidate(),
  })
}

export function useMergeCanteenStore() {
  const invalidate = useCanteenInvalidate()
  return useMutation({
    mutationFn: ({ storeID, targetID }: { storeID: number; targetID: number }) =>
      mergeCanteenStore(storeID, targetID),
    onSuccess: () => void invalidate(),
  })
}
