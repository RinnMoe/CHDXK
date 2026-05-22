import { RiThumbUpLine, RiTimeLine } from "@remixicon/react"
import type { FilterItem } from "@/api/types"
import type { ReviewListFilter } from "@/api/review"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/utils"

const ALL = "all"

type ReviewSort = NonNullable<ReviewListFilter["order_by"]>

interface CourseReviewFilterValue {
  semester?: string
  rating?: number
  orderBy: ReviewSort
}

interface CourseReviewFiltersProps {
  semesters: FilterItem[]
  ratings: FilterItem[]
  total: number
  value: CourseReviewFilterValue
  onChange: (next: Partial<CourseReviewFilterValue>) => void
}

function OptionLabel({ label, count }: { label: string; count: number }) {
  return (
    <span className="flex min-w-0 items-center gap-1">
      <span className="truncate">{label}</span>
      <span className="shrink-0 text-xs text-muted-foreground">
        （{count}条点评）
      </span>
    </span>
  )
}

export function CourseReviewFilters({
  semesters,
  ratings,
  total,
  value,
  onChange,
}: CourseReviewFiltersProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Select
        value={value.semester ?? ALL}
        onValueChange={(semester) =>
          onChange({ semester: semester === ALL ? undefined : semester })
        }
      >
        <SelectTrigger size="sm" className="w-[220px]">
          <SelectValue placeholder="学期" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>
            <OptionLabel label="全部学期" count={total} />
          </SelectItem>
          {semesters.map((semester) => (
            <SelectItem key={semester.name} value={semester.name}>
              <OptionLabel label={semester.name} count={semester.count} />
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={value.rating ? String(value.rating) : ALL}
        onValueChange={(rating) =>
          onChange({ rating: rating === ALL ? undefined : Number(rating) })
        }
      >
        <SelectTrigger size="sm" className="w-[170px]">
          <SelectValue placeholder="评分" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>
            <OptionLabel label="全部评分" count={total} />
          </SelectItem>
          {ratings.map((rating) => (
            <SelectItem key={rating.name} value={rating.name}>
              <OptionLabel label={`${rating.name} 分`} count={rating.count} />
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <div className="flex items-center gap-1">
        <span className="px-1 text-sm text-muted-foreground">排序</span>
        <div
          role="tablist"
          aria-label="评价排序"
          className="inline-flex h-8 items-center rounded-md bg-muted p-0.5"
        >
          <button
            type="button"
            role="tab"
            aria-selected={value.orderBy === "created_at"}
            className={cn(
              "inline-flex h-7 items-center gap-1 rounded-sm px-2.5 text-sm font-normal transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
              value.orderBy === "created_at"
                ? "bg-background text-foreground shadow-xs"
                : "text-muted-foreground hover:text-foreground"
            )}
            onClick={() => onChange({ orderBy: "created_at" })}
          >
            <RiTimeLine className="size-3" data-icon="inline-start" />
            最新
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={value.orderBy === "like_count"}
            className={cn(
              "inline-flex h-7 items-center gap-1 rounded-sm px-2.5 text-sm font-normal transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
              value.orderBy === "like_count"
                ? "bg-background text-foreground shadow-xs"
                : "text-muted-foreground hover:text-foreground"
            )}
            onClick={() => onChange({ orderBy: "like_count" })}
          >
            <RiThumbUpLine className="size-3" data-icon="inline-start" />
            获赞最多
          </button>
        </div>
      </div>
    </div>
  )
}
