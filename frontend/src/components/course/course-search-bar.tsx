import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import { Input } from "@/components/ui/input"

export function CourseSearchBar() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [value, setValue] = useState(searchParams.get("code") ?? "")

  useEffect(() => {
    const handler = setTimeout(() => {
      const current = searchParams.get("code") ?? ""
      if (value === current) return
      const next = new URLSearchParams(searchParams)
      if (value) {
        next.set("code", value)
      } else {
        next.delete("code")
      }
      next.delete("page")
      setSearchParams(next)
    }, 300)
    return () => clearTimeout(handler)
  }, [value, searchParams, setSearchParams])

  return (
    <Input
      placeholder="搜索课程代码或名称..."
      value={value}
      onChange={(e) => setValue(e.target.value)}
      className="max-w-md"
    />
  )
}
