import { RiThumbUpLine, RiTimeLine } from "@remixicon/react"
import type { FilterItem } from "@/api/types"
import type { ReviewListFilter } from "@/api/review"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

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
        <Button
          type="button"
          size="sm"
          variant={value.orderBy === "created_at" ? "secondary" : "outline"}
          onClick={() => onChange({ orderBy: "created_at" })}
        >
          <RiTimeLine data-icon="inline-start" />
          最新
        </Button>
        <Button
          type="button"
          size="sm"
          variant={value.orderBy === "like_count" ? "secondary" : "outline"}
          onClick={() => onChange({ orderBy: "like_count" })}
        >
          <RiThumbUpLine data-icon="inline-start" />
          获赞最多
        </Button>
      </div>
    </div>
  )
}
