import { useState } from "react"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { RiFilterLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
const routeApi = getRouteApi("/app/course")
type CourseSearch = ReturnType<typeof routeApi.useSearch>
type FilterKey = keyof Pick<
  CourseSearch,
  | "categories"
  | "credit"
  | "department"
  | "language"
  | "order_by"
  | "target_years"
>

interface CourseFiltersProps {
  filters: CourseFiltersDTO
}

export function CourseFilters({ filters }: CourseFiltersProps) {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/course" })
  const [mobileOpen, setMobileOpen] = useState(false)
  const orderBy = search.order_by ?? "rating_score"

  function updateFilter(key: FilterKey, value: string | string[] | null) {
    let nextValue: string | string[] | number | undefined
    if (
      value === null ||
      value === "" ||
      value === ALL ||
      (Array.isArray(value) && value.length === 0)
    ) {
      nextValue = undefined
    } else if (Array.isArray(value)) {
      nextValue = value
    } else if (key === "credit") {
      const credit = Number(value)
      nextValue = Number.isFinite(credit) ? credit : undefined
    } else {
      nextValue = value
    }

    void navigate({
      search: (prev) => ({
        ...prev,
        [key]: nextValue,
        page: 1,
      }),
      resetScroll: false,
    })
  }

  function toggleMulti(key: "categories" | "target_years", value: string) {
    const cur = search[key] ?? []
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
          value={orderBy}
          onValueChange={(v) => updateFilter("order_by", v)}
        >
          <TabsList>
            <TabsTrigger value="rating_score">综合评分</TabsTrigger>
            <TabsTrigger value="rating_count">点评数量</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {filters.categories && (
        <FilterSelectGroup
          label="课程类别"
          items={filters.categories}
          paramKey="categories"
          selected={search.categories?.[0] ?? null}
          onChange={updateFilter}
        />
      )}

      {filters.credits && (
        <FilterSelectGroup
          label="学分"
          items={filters.credits}
          paramKey="credit"
          selected={search.credit === undefined ? null : String(search.credit)}
          onChange={updateFilter}
        />
      )}

      {filters.languages && (
        <FilterSelectGroup
          label="授课语言"
          items={filters.languages}
          paramKey="language"
          selected={search.language ?? null}
          onChange={updateFilter}
        />
      )}

      {filters.target_years && (
        <FilterCheckGroup
          label="目标年级"
          items={filters.target_years}
          paramKey="target_years"
          selected={search.target_years ?? []}
          onToggle={toggleMulti}
        />
      )}

      {filters.departments && (
        <FilterSelectGroup
          label="开课单位"
          items={filters.departments}
          paramKey="department"
          selected={search.department ?? null}
          onChange={updateFilter}
        />
      )}

      <Button
        variant="outline"
        className="w-full"
        onClick={() => {
          void navigate({
            search: {
              q: search.q,
              department: undefined,
              language: undefined,
              categories: undefined,
              target_years: undefined,
              credit: undefined,
              order_by: undefined,
              page: 1,
            },
            resetScroll: false,
          })
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

interface FilterSelectGroupProps {
  label: string
  paramKey: FilterKey
  items?: FilterItem[]
  selected: string | null
  onChange: (key: FilterKey, value: string | null) => void
}

function FilterSelectGroup({
  label,
  paramKey,
  items,
  selected,
  onChange,
}: FilterSelectGroupProps) {
  if (!items) return null
  const itemClassName =
    "px-2 [&>span:first-child]:hidden [&>span:last-child]:min-w-0 [&>span:last-child]:flex-1"

  return (
    <div className="space-y-3">
      <Label>{label}</Label>
      <Select
        value={selected ?? ALL}
        onValueChange={(value) => onChange(paramKey, value)}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={`全部${label}`} />
        </SelectTrigger>
        <SelectContent className="max-w-80">
          <SelectItem value={ALL} className={itemClassName}>
            全部{label}
          </SelectItem>
          {items.map((item) => (
            <SelectItem
              key={item.name}
              value={item.name}
              className={itemClassName}
            >
              <span className="flex w-full min-w-0 items-center justify-between gap-3">
                <span className="min-w-0 truncate">{item.name}</span>
                <span className="shrink-0 text-muted-foreground">
                  {item.count}
                </span>
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}

interface FilterCheckGroupProps {
  label: string
  paramKey: "categories" | "target_years"
  items?: FilterItem[]
  selected: string[]
  onToggle: (key: "categories" | "target_years", value: string) => void
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
