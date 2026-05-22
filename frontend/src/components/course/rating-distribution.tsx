import type { RatingInfoDTO } from "@/api/course"
import { RatingDisplay } from "./rating-display"

interface RatingDistributionProps {
  rating: RatingInfoDTO
}

export function RatingDistribution({ rating }: RatingDistributionProps) {
  const max = Math.max(...rating.distribution, 1)

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-8">
      <div className="flex flex-col items-center gap-1 sm:items-start">
        <div className="text-3xl font-bold">
          {rating.avg > 0 ? rating.avg.toFixed(1) : "—"}
        </div>
        <RatingDisplay rating={rating} />
      </div>

      <div className="flex-1 space-y-1.5 max-w-xs">
        {[5, 4, 3, 2, 1].map((star) => {
          const count = rating.distribution[star - 1]
          const pct = (count / max) * 100
          return (
            <div key={star} className="flex items-center gap-2 text-xs">
              <span className="w-6 text-muted-foreground">{star} 星</span>
              <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full bg-yellow-400"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <span className="w-8 text-right text-muted-foreground tabular-nums">
                {count}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
