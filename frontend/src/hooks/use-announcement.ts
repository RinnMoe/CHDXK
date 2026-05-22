import { useQuery } from "@tanstack/react-query"
import { listAnnouncements } from "@/api/announcement"

export function useAnnouncements() {
  return useQuery({
    queryKey: ["announcements"],
    queryFn: listAnnouncements,
    staleTime: 1000 * 60 * 5,
  })
}
