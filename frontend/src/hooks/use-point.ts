import { keepPreviousData, useQuery } from "@tanstack/react-query"
import {
  getUserPoints,
  type PointRecordListFilter,
} from "@/api/point"

export function useUserPoints(
  userID: number,
  filter: PointRecordListFilter = {}
) {
  return useQuery({
    queryKey: ["points", userID, filter],
    queryFn: () => getUserPoints(userID, filter),
    enabled: !!userID,
    placeholderData: keepPreviousData,
  })
}
