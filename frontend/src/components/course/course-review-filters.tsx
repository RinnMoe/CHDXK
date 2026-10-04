import { copy } from "@/lib/copy"
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
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"

const ALL = "all"

type ReviewSort = NonNullable<ReviewListFilter["order_by"]>

interface CourseReviewFilterValue {
  semester?: string
  rating?: number
  orderBy: ReviewSort
}

interface CourseReviewFiltersProps {
  semesters?: FilterItem[]
  ratings?: FilterItem[]
  value: CourseReviewFilterValue
  onChange: (next: Partial<CourseReviewFilterValue>) => void
}

function OptionLabel({ label, count }: { label: string; count: number }) {
  return (
    <span className="flex min-w-0 items-center gap-1">
      <span className="truncate">{label}</span>
      <span className="shrink-0 text-sm text-muted-foreground">
        （{count}
        {copy("copy.user.components.course.course-review-filters.9f838b6a529b")}
      </span>
    </span>
  )
}

export function CourseReviewFilters({
  semesters,
  ratings,
  value,
  onChange,
}: CourseReviewFiltersProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {semesters && (
        <Select
          value={value.semester ?? ALL}
          onValueChange={(semester) =>
            onChange({ semester: semester === ALL ? undefined : semester })
          }
        >
          <SelectTrigger size="sm" className="w-55">
            <SelectValue
              placeholder={copy(
                "copy.user.components.course.course-review-filters.79a4ef019e9c"
              )}
            />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>
              {copy(
                "copy.user.components.course.course-review-filters.a0670d57e5bf"
              )}
            </SelectItem>
            {semesters.map((semester) => (
              <SelectItem key={semester.name} value={semester.name}>
                <OptionLabel label={semester.name} count={semester.count} />
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      {ratings && (
        <Select
          value={value.rating ? String(value.rating) : ALL}
          onValueChange={(rating) =>
            onChange({ rating: rating === ALL ? undefined : Number(rating) })
          }
        >
          <SelectTrigger size="sm" className="w-42.5">
            <SelectValue
              placeholder={copy(
                "copy.user.components.course.course-review-filters.96fff2e26c8d"
              )}
            />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>
              {copy(
                "copy.user.components.course.course-review-filters.ee9dc66201f7"
              )}
            </SelectItem>
            {ratings.map((rating) => (
              <SelectItem key={rating.name} value={rating.name}>
                <OptionLabel
                  label={copy(
                    "copy.user.components.course.course-review-filters.180ba6fb8c00",
                    { v0: rating.name }
                  )}
                  count={rating.count}
                />
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      <div className="flex items-center gap-1">
        <span className="px-1 text-sm text-muted-foreground">
          {copy(
            "copy.user.components.course.course-review-filters.a96c9a854190"
          )}
        </span>
        <Tabs
          value={value.orderBy}
          onValueChange={(orderBy) =>
            onChange({ orderBy: orderBy as ReviewSort })
          }
          className="gap-0"
        >
          <TabsList>
            <TabsTrigger value="updated_at" className="gap-1">
              <RiTimeLine className="size-3" data-icon="inline-start" />
              {copy(
                "copy.user.components.course.course-review-filters.569af1a47dbd"
              )}
            </TabsTrigger>
            <TabsTrigger value="like_count" className="gap-1">
              <RiThumbUpLine className="size-3" data-icon="inline-start" />
              {copy(
                "copy.user.components.course.course-review-filters.4eecf5a9a9da"
              )}
            </TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
    </div>
  )
}
