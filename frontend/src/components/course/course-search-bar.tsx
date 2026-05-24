import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import { useDebounce } from "use-debounce"
import { Input } from "@/components/ui/input"

const SEARCH_DEBOUNCE_MS = 250

export function CourseSearchBar() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [value, setValue] = useState((searchParams.get("q") ?? "").trim())
  const [debouncedValue] = useDebounce(value, SEARCH_DEBOUNCE_MS)

  useEffect(() => {
    const nextQ = debouncedValue.trim()
    const current = (searchParams.get("q") ?? "").trim()
    if (nextQ === current) return
    const next = new URLSearchParams(searchParams)
    if (nextQ) {
      next.set("q", nextQ)
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
