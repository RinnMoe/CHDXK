import { useState } from "react"
import { useSearchParams } from "react-router-dom"
import { RiFilterLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
import type { CourseFilters as CourseFiltersDTO } from "@/api/course"

const ALL = "__all__"

interface CourseFiltersProps {
  filters: CourseFiltersDTO
}

export function CourseFilters({ filters }: CourseFiltersProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const [mobileOpen, setMobileOpen] = useState(false)

  function updateFilter(key: string, value: string | string[] | null) {
    const next = new URLSearchParams(searchParams)
    if (
      value === null ||
      value === "" ||
      value === ALL ||
      (Array.isArray(value) && value.length === 0)
    ) {
      next.delete(key)
    } else if (Array.isArray(value)) {
      next.delete(key)
      value.forEach((v) => next.append(key, v))
    } else {
      next.set(key, value)
    }
    next.delete("page")
    setSearchParams(next)
  }

  function toggleMulti(key: string, value: string) {
    const cur = searchParams.getAll(key)
    const next = cur.includes(value)
      ? cur.filter((v) => v !== value)
      : [...cur, value]
    updateFilter(key, next)
  }

  const content = (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label>排序</Label>
        <Select
          value={searchParams.get("order_by") ?? ALL}
          onValueChange={(v) => updateFilter("order_by", v)}
        >
          <SelectTrigger>
            <SelectValue placeholder="默认排序" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>默认排序</SelectItem>
            <SelectItem value="rating_count">评价数量</SelectItem>
            <SelectItem value="rating_avg">平均评分</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>授课语言</Label>
        <Select
          value={searchParams.get("language") ?? ALL}
          onValueChange={(v) => updateFilter("language", v)}
        >
          <SelectTrigger>
            <SelectValue placeholder="全部语言" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>全部语言</SelectItem>
            <SelectItem value="中文">中文</SelectItem>
            <SelectItem value="英文">英文</SelectItem>
            <SelectItem value="双语">双语</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Separator />

      <FilterCheckGroup
        label="学院"
        items={filters.departments}
        paramKey="department"
        selected={searchParams.getAll("department")}
        onToggle={toggleMulti}
      />

      <Separator />

      <FilterCheckGroup
        label="课程类别"
        items={filters.categories}
        paramKey="categories"
        selected={searchParams.getAll("categories")}
        onToggle={toggleMulti}
      />

      <Separator />

      <FilterCheckGroup
        label="目标年级"
        items={filters.target_years}
        paramKey="target_years"
        selected={searchParams.getAll("target_years")}
        onToggle={toggleMulti}
      />

      <Separator />

      <div className="space-y-3">
        <Label>学分</Label>
        <div className="flex gap-2">
          {filters.credits.map((c) => (
            <Button
              key={c.name}
              size="sm"
              variant={
                searchParams.getAll("credit").includes(c.name) ? "default" : "outline"
              }
              onClick={() => toggleMulti("credit", c.name)}
            >
              {c.name}
            </Button>
          ))}
        </div>
      </div>

      <Button
        variant="ghost"
        className="w-full"
        onClick={() => {
          const next = new URLSearchParams()
          const code = searchParams.get("code")
          if (code) next.set("code", code)
          setSearchParams(next)
        }}
      >
        清除筛选
      </Button>
    </div>
  )

  return (
    <>
      <aside className="hidden lg:block w-56 shrink-0">
        <div className="sticky top-20">{content}</div>
      </aside>
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
              <SheetTitle>筛选课程</SheetTitle>
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
  selected: string[]
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
          const checked = selected.includes(item.name)
          return (
            <div key={item.name} className="flex items-center gap-2">
              <Checkbox
                id={id}
                checked={checked}
                onCheckedChange={() => onToggle(paramKey, item.name)}
              />
              <Label
                htmlFor={id}
                className="text-sm font-normal cursor-pointer flex-1 truncate"
              >
                {item.name}
              </Label>
              <span className="text-xs text-muted-foreground shrink-0">
                {item.count}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
