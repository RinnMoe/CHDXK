import { ReviewCard } from "./review-card"
import { Skeleton } from "@/components/ui/skeleton"
import type { ReviewDTO } from "@/api/review"

interface ReviewListProps {
  reviews: ReviewDTO[]
  isLoading?: boolean
  showCourse?: boolean
  emptyText?: string
}

function ReviewCardSkeleton() {
  return (
    <div className="rounded-lg border p-4 space-y-3">
      <div className="flex gap-2">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-12" />
      </div>
      <Skeleton className="h-3 w-full" />
      <Skeleton className="h-3 w-5/6" />
      <Skeleton className="h-3 w-2/3" />
    </div>
  )
}

export function ReviewList({
  reviews,
  isLoading,
  showCourse,
  emptyText = "暂无评价",
}: ReviewListProps) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <ReviewCardSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (reviews.length === 0) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm text-muted-foreground">{emptyText}</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {reviews.map((review) => (
        <ReviewCard key={review.id} review={review} showCourse={showCourse} />
      ))}
    </div>
  )
}
