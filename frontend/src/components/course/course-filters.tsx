import { useState } from "react"
import { useSearchParams } from "react-router-dom"
import { RiFilterLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
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
import type { FilterItem } from "@/api/types"
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
        <Tabs
          value={searchParams.get("order_by") ?? ALL}
          onValueChange={(v) => updateFilter("order_by", v)}
        >
          <TabsList>
            <TabsTrigger value={ALL}>默认</TabsTrigger>
            <TabsTrigger value="rating_count">点评数量</TabsTrigger>
            <TabsTrigger value="rating_avg">平均评分</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {filters.languages && (
        <>
          <FilterCheckGroup
            label="授课语言"
            items={filters.languages}
            paramKey="language"
            selected={searchParams.getAll("language")}
            onToggle={toggleMulti}
          />
          <Separator />
        </>
      )}

      {filters.departments && (
        <>
          <FilterCheckGroup
            label="学院"
            items={filters.departments}
            paramKey="department"
            selected={searchParams.getAll("department")}
            onToggle={toggleMulti}
          />
          <Separator />
        </>
      )}

      {filters.categories && (
        <>
          <FilterCheckGroup
            label="课程类别"
            items={filters.categories}
            paramKey="categories"
            selected={searchParams.getAll("categories")}
            onToggle={toggleMulti}
          />
          <Separator />
        </>
      )}

      {filters.target_years && (
        <>
          <FilterCheckGroup
            label="目标年级"
            items={filters.target_years}
            paramKey="target_years"
            selected={searchParams.getAll("target_years")}
            onToggle={toggleMulti}
          />
          <Separator />
        </>
      )}

      {filters.credits && (
        <FilterCheckGroup
          label="学分"
          items={filters.credits}
          paramKey="credit"
          selected={searchParams.getAll("credit")}
          onToggle={toggleMulti}
        />
      )}

      <Button
        variant="outline"
        className="w-full"
        onClick={() => {
          const next = new URLSearchParams()
          const q = searchParams.get("q")?.trim()
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
  items?: FilterItem[]
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
  if (!items) return null

  return (
    <div className="space-y-3">
      <Label>{label}</Label>
      <div className="space-y-2">
        {items.map((item) => {
          const id = `${paramKey}-${item.name}`
          const checked = selected.includes(item.name)
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
