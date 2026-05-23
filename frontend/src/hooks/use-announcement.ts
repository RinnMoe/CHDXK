import { useQuery } from "@tanstack/react-query"
import { listAnnouncements } from "@/api/announcement"

export function useAnnouncements(enabled = true) {
  return useQuery({
    queryKey: ["announcements"],
    queryFn: listAnnouncements,
    enabled,
    staleTime: 1000 * 60 * 5,
  })
}
