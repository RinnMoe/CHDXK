import { Button } from "@/components/ui/button"
import { copy } from "@/lib/copy"
import { Input } from "@/components/ui/input"
import { TableHead, TableHeader, TableRow } from "@/components/ui/table"

interface Props {
  filters: {
    code: string
    name: string
    teacher: string
    department: string
    lesson_no: string
    min_sections: string
    max_sections: string
  }
  sort: "code" | "name" | "teacher" | "department" | "sections"
  direction: "asc" | "desc"
  onFilterChange: (key: keyof Props["filters"], value: string) => void
  onSortChange: (key: Props["sort"]) => void
}

export function ManagedCourseTableHeader({
  filters,
  sort,
  direction,
  onFilterChange,
  onSortChange,
}: Props) {
  return (
    <TableHeader>
      <TableRow>
        {(
          [
            ["code", "课程代码"],
            ["name", "课程名称"],
            ["teacher", "主讲教师"],
            ["department", "开课单位"],
            ["sections", "教学班数量"],
          ] as const
        ).map(([key, label]) => (
          <TableHead
            key={key}
            aria-sort={
              sort === key
                ? direction === "asc"
                  ? "ascending"
                  : "descending"
                : "none"
            }
          >
            <button
              className="inline-flex items-center gap-1 py-2 whitespace-nowrap hover:text-foreground"
              onClick={() => onSortChange(key)}
              aria-label={`按${label}排序`}
            >
              {label}
              <span aria-hidden="true">
                {sort === key ? (direction === "asc" ? "↑" : "↓") : "↕"}
              </span>
            </button>
          </TableHead>
        ))}
        <TableHead className="text-right">
          {copy(
            "copy.admin.components.admin.course-management-tab.ed31fbb483ee"
          )}
        </TableHead>
      </TableRow>
      <TableRow>
        {(
          [
            ["code", "课程代码"],
            ["name", "课程名称"],
            ["teacher", "主讲教师"],
            ["department", "开课单位"],
          ] as const
        ).map(([key, label]) => (
          <TableHead key={key} className="py-2">
            <Input
              className="min-w-28"
              aria-label={`筛选${label}`}
              value={filters[key]}
              placeholder={`搜索${label}`}
              onChange={(event) => onFilterChange(key, event.target.value)}
            />
          </TableHead>
        ))}
        <TableHead className="py-2">
          <div className="flex min-w-36 items-center gap-1">
            <Input
              type="number"
              min={0}
              step={1}
              aria-label="最少教学班数量"
              value={filters.min_sections}
              placeholder="最少"
              onChange={(event) =>
                onFilterChange("min_sections", event.target.value)
              }
            />
            <span>—</span>
            <Input
              type="number"
              min={0}
              step={1}
              aria-label="最多教学班数量"
              value={filters.max_sections}
              placeholder="最多"
              onChange={(event) =>
                onFilterChange("max_sections", event.target.value)
              }
            />
          </div>
        </TableHead>
        <TableHead />
      </TableRow>
    </TableHeader>
  )
}

export function ManagedCoursePagination({
  total,
  page,
  pageCount,
  onPageChange,
}: {
  total: number
  page: number
  pageCount: number
  onPageChange: (page: number) => void
}) {
  return (
    <div className="flex items-center justify-between text-sm text-muted-foreground">
      <span>
        {copy("copy.admin.components.admin.course-management-tab.2428420e407f")}
        {total}
        {copy("copy.admin.components.admin.course-management-tab.88165354d9d3")}
      </span>
      <div className="flex items-center gap-2">
        <span>
          {copy(
            "copy.admin.components.admin.course-management-tab.9552fe3bd366"
          )}
          {page} / {pageCount}
          {copy(
            "copy.admin.components.admin.course-management-tab.85ca81cf6abd"
          )}
        </span>
        <Button
          size="sm"
          variant="outline"
          disabled={page <= 1}
          onClick={() => onPageChange(Math.max(1, page - 1))}
        >
          {copy(
            "copy.admin.components.admin.course-management-tab.c9b9ae7a6144"
          )}
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={page >= pageCount}
          onClick={() => onPageChange(Math.min(pageCount, page + 1))}
        >
          {copy(
            "copy.admin.components.admin.course-management-tab.8a8542f69648"
          )}
        </Button>
      </div>
    </div>
  )
}
