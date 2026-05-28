import { useEffect, useState } from "react"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useDebounce } from "use-debounce"
import { Input } from "@/components/ui/input"

const SEARCH_DEBOUNCE_MS = 250
const routeApi = getRouteApi("/app/course")

export function CourseSearchBar() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/course" })
  const [value, setValue] = useState((search.q ?? "").trim())
  const [debouncedValue] = useDebounce(value, SEARCH_DEBOUNCE_MS)

  useEffect(() => {
    const nextQ = debouncedValue.trim()
    const current = (search.q ?? "").trim()
    if (nextQ === current) return

    void navigate({
      search: (prev) => ({
        ...prev,
        q: nextQ || undefined,
        page: 1,
      }),
      replace: true,
      resetScroll: false,
    })
  }, [debouncedValue, navigate, search.q])

  return (
    <Input
      key={search.q ?? ""}
      placeholder="搜索课程名、课程号或教师名..."
      value={value}
      onChange={(e) => setValue(e.target.value)}
    />
  )
}
