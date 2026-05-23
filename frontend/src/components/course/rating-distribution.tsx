import type { RatingInfoDTO } from "@/api/course"
import { RiStarFill, RiStarLine } from "@remixicon/react"

interface RatingDistributionProps {
  rating: RatingInfoDTO
}

export function RatingDistribution({ rating }: RatingDistributionProps) {
  const total = Math.max(rating.count, 1)
  const filled = Math.round(rating.avg)
  const stars = Array.from({ length: 5 }, (_, i) => i < filled)

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-8 md:flex-col md:items-stretch md:gap-3 2xl:flex-row 2xl:items-center 2xl:gap-8">
      <div className="flex items-center justify-center gap-6">
        <div className="text-3xl font-bold tabular-nums">
          {rating.avg > 0 ? rating.avg.toFixed(1) : "—"}
        </div>
        <div className="flex flex-col items-center gap-1">
          <div className="flex">
            {stars.map((on, index) =>
              on ? (
                <RiStarFill key={index} className="size-3.5 text-yellow-400" />
              ) : (
                <RiStarLine
                  key={index}
                  className="size-3.5 text-muted-foreground/40"
                />
              )
            )}
          </div>
          <span className="text-center text-sm text-muted-foreground">
            {rating.count}条点评
          </span>
        </div>
      </div>

      <div className="max-w-xs flex-1 space-y-1.5">
        {[5, 4, 3, 2, 1].map((star) => {
          const count = rating.distribution[star - 1]
          const pct = (count / total) * 100
          return (
            <div key={star} className="flex items-center gap-2 text-sm">
              <span className="w-8 shrink-0 text-muted-foreground">
                {star} 星
              </span>
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
