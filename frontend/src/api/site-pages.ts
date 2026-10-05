import { apiClient } from "./client"
import { useQuery } from "@tanstack/react-query"

export interface SitePage {
  slug: string
  title: string
  content: string
  published: boolean
  footer: boolean
}
export function useSitePages() {
  return useQuery({
    queryKey: ["site-pages"],
    queryFn: () => apiClient<SitePage[]>("/api/site-pages"),
    staleTime: 60000,
  })
}
