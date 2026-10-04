import { copy } from "@/lib/copy"
import { useState } from "react"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { RiFilterLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
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
  | "semester"
  | "campus"
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
    } else if (key === "categories" || key === "target_years") {
      nextValue = [value]
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

  const content = (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label>
          {copy("copy.user.components.course.course-filters.a96c9a854190")}
        </Label>
        <Tabs
          value={orderBy}
          onValueChange={(v) => updateFilter("order_by", v)}
        >
          <TabsList>
            <TabsTrigger value="rating_score">
              {copy("copy.user.components.course.course-filters.6408c2c76e3b")}
            </TabsTrigger>
            <TabsTrigger value="rating_count">
              {copy("copy.user.components.course.course-filters.ed4a6ed53604")}
            </TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {filters.semesters && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.32090f3541c7"
          )}
          items={filters.semesters}
          paramKey="semester"
          selected={search.semester ?? null}
          onChange={updateFilter}
        />
      )}
      {filters.campuses && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.406c59bfe6b8"
          )}
          items={filters.campuses}
          paramKey="campus"
          selected={search.campus ?? null}
          onChange={updateFilter}
        />
      )}
      {filters.categories && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.66e8e3adf36b"
          )}
          items={filters.categories}
          paramKey="categories"
          selected={search.categories?.[0] ?? null}
          onChange={updateFilter}
        />
      )}

      {filters.credits && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.f9d072c4804c"
          )}
          items={filters.credits}
          paramKey="credit"
          selected={search.credit === undefined ? null : String(search.credit)}
          onChange={updateFilter}
        />
      )}

      {filters.languages && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.2cbd78206ec3"
          )}
          items={filters.languages}
          paramKey="language"
          selected={search.language ?? null}
          onChange={updateFilter}
        />
      )}

      {filters.target_years && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.cbd392ead6c6"
          )}
          items={filters.target_years}
          paramKey="target_years"
          selected={search.target_years?.[0] ?? null}
          onChange={updateFilter}
        />
      )}

      {filters.departments && (
        <FilterSelectGroup
          label={copy(
            "copy.user.components.course.course-filters.6322fa286117"
          )}
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
              semester: undefined,
              campus: undefined,
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
        {copy("copy.user.components.course.course-filters.657d9cbf45ec")}
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
              {copy("copy.user.components.course.course-filters.b5f15473fdc0")}
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-72">
            <SheetHeader>
              <SheetTitle>
                {copy(
                  "copy.user.components.course.course-filters.647f8a821b2b"
                )}
              </SheetTitle>
            </SheetHeader>
            <ScrollArea className="mt-4 h-[calc(100vh-6rem)] px-4">
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
  const selectedItem = selected
    ? items.find((item) => item.name === selected)
    : undefined
  const itemClassName =
    "px-2 [&>span:first-child]:hidden [&>span:last-child]:min-w-0 [&>span:last-child]:flex-1"

  return (
    <div className="space-y-3">
      <Label>{label}</Label>
      <Select
        value={selected ?? ALL}
        onValueChange={(value) => onChange(paramKey, value)}
      >
        <SelectTrigger className="w-full [&>[data-slot=select-value]]:min-w-0 [&>[data-slot=select-value]]:flex-1 [&>[data-slot=select-value]]:gap-0 [&>[data-slot=select-value]>span]:w-full">
          <SelectValue>
            {selectedItem ? (
              <FilterOptionLabel item={selectedItem} />
            ) : (
              (selected ??
              copy("copy.user.components.course.course-filters.d50f21139128", {
                v0: label,
              }))
            )}
          </SelectValue>
        </SelectTrigger>
        <SelectContent className="max-w-80">
          <SelectItem value={ALL} className={itemClassName}>
            {copy("copy.user.components.course.course-filters.5c55a67935af")}
            {label}
          </SelectItem>
          {items.map((item) => (
            <SelectItem
              key={item.name}
              value={item.name}
              className={itemClassName}
            >
              <FilterOptionLabel item={item} />
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}

function FilterOptionLabel({ item }: { item: FilterItem }) {
  return (
    <span className="flex w-full min-w-0 items-center justify-between gap-3">
      <span className="min-w-0 truncate">{item.name}</span>
      <span className="shrink-0 text-muted-foreground">{item.count}</span>
    </span>
  )
}
