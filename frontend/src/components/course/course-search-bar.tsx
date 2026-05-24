import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import { useDebounce } from "use-debounce"
import { Input } from "@/components/ui/input"

const SEARCH_DEBOUNCE_MS = 250

export function CourseSearchBar() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [value, setValue] = useState(searchParams.get("q") ?? "")
  const [debouncedValue] = useDebounce(value, SEARCH_DEBOUNCE_MS)

  useEffect(() => {
    const current = searchParams.get("q") ?? ""
    if (debouncedValue === current) return
    const next = new URLSearchParams(searchParams)
    if (debouncedValue) {
      next.set("q", debouncedValue)
    } else {
      next.delete("q")
    }
    next.delete("page")
    setSearchParams(next)
  }, [debouncedValue, searchParams, setSearchParams])

  return (
    <Input
      placeholder="搜索课程、代码或教师..."
      value={value}
      onChange={(e) => setValue(e.target.value)}
    />
  )
}
