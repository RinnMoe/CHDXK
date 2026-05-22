import type { RatingInfoDTO } from "@/api/course"
import { RatingDisplay } from "./rating-display"

interface RatingDistributionProps {
  rating: RatingInfoDTO
}

export function RatingDistribution({ rating }: RatingDistributionProps) {
  const total = Math.max(rating.count, 1)

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-8">
      <div className="flex flex-col items-center gap-1 sm:items-start">
        <div className="text-3xl font-bold">
          {rating.avg > 0 ? rating.avg.toFixed(1) : "—"}
        </div>
        <RatingDisplay rating={rating} />
      </div>

      <div className="max-w-xs flex-1 space-y-1.5">
        {[5, 4, 3, 2, 1].map((star) => {
          const count = rating.distribution[star - 1]
          const pct = (count / total) * 100
          return (
            <div key={star} className="flex items-center gap-2 text-xs">
              <span className="w-6 text-muted-foreground">{star} 星</span>
              <div className="h-2 flex-1 overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full bg-yellow-400"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <span className="w-16 text-right text-muted-foreground tabular-nums">
                {count}条点评
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
