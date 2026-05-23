import { useState } from "react"
import { useSearchParams } from "react-router-dom"
import { RiFilterLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { ScrollArea } from "@/components/ui/scroll-area"
import type { TeacherFilters as TeacherFiltersDTO } from "@/api/teacher"

interface TeacherFiltersProps {
  filters: TeacherFiltersDTO
}

export function TeacherFilters({ filters }: TeacherFiltersProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const [mobileOpen, setMobileOpen] = useState(false)

  function updateFilter(key: string, value: string | null) {
    const next = new URLSearchParams(searchParams)
    if (!value) next.delete(key)
    else next.set(key, value)
    next.delete("page")
    setSearchParams(next)
  }

  function toggleSingle(key: string, value: string) {
    updateFilter(key, searchParams.get(key) === value ? null : value)
  }

  const content = (
    <div className="space-y-6">
      <FilterCheckGroup
        label="学院"
        items={filters.departments}
        paramKey="department"
        selected={searchParams.get("department")}
        onToggle={toggleSingle}
      />

      <Separator />

      <FilterCheckGroup
        label="职称"
        items={filters.titles}
        paramKey="title"
        selected={searchParams.get("title")}
        onToggle={toggleSingle}
      />

      <Button
        variant="outline"
        className="w-full"
        onClick={() => {
          const next = new URLSearchParams()
          const q = searchParams.get("q")
          if (q) next.set("q", q)
          setSearchParams(next)
        }}
      >
        清除筛选
      </Button>
    </div>
  )

  return (
    <>
      <aside className="hidden lg:block">{content}</aside>
      <div className="lg:hidden">
        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <SheetTrigger asChild>
            <Button variant="outline" size="sm">
              <RiFilterLine data-icon="inline-start" />
              筛选
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-72">
            <SheetHeader>
              <SheetTitle>筛选教师</SheetTitle>
            </SheetHeader>
            <ScrollArea className="mt-4 h-[calc(100vh-6rem)] pr-4 pl-4">
              {content}
            </ScrollArea>
          </SheetContent>
        </Sheet>
      </div>
    </>
  )
}

interface FilterCheckGroupProps {
  label: string
  paramKey: string
  items: { name: string; count: number }[]
  selected: string | null
  onToggle: (key: string, value: string) => void
}

function FilterCheckGroup({
  label,
  paramKey,
  items,
  selected,
  onToggle,
}: FilterCheckGroupProps) {
  return (
    <div className="space-y-3">
      <Label>{label}</Label>
      <div className="space-y-2">
        {items.map((item) => {
          const id = `${paramKey}-${item.name}`
          const checked = selected === item.name
          return (
            <div key={item.name} className="flex items-start gap-2">
              <Checkbox
                id={id}
                checked={checked}
                className="mt-0.5"
                onCheckedChange={() => onToggle(paramKey, item.name)}
              />
              <Label
                htmlFor={id}
                className="block min-w-0 flex-1 cursor-pointer text-sm leading-snug font-normal break-all whitespace-normal"
              >
                {item.name}
              </Label>
              <span className="mt-0.5 shrink-0 text-sm text-muted-foreground">
                {item.count}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
