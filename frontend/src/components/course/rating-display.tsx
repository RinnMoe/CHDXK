import { RiStarFill, RiStarLine } from "@remixicon/react"
import type { RatingInfoDTO } from "@/api/course"

interface RatingDisplayProps {
  rating: RatingInfoDTO
  size?: "sm" | "md"
}

export function RatingDisplay({ rating, size = "md" }: RatingDisplayProps) {
  const sizeClass = size === "sm" ? "size-3" : "size-3.5"
  const filled = Math.round(rating.avg)
  const stars = Array.from({ length: 5 }, (_, i) => i < filled)

  return (
    <div className="flex items-center gap-1.5">
      <div className="flex">
        {stars.map((on, i) =>
          on ? (
            <RiStarFill key={i} className={`${sizeClass} text-yellow-400`} />
          ) : (
            <RiStarLine
              key={i}
              className={`${sizeClass} text-muted-foreground/40`}
            />
          )
        )}
      </div>
      {rating.avg > 0 && (
        <span className="text-sm font-medium">{rating.avg.toFixed(1)}</span>
      )}
      <span className="text-xs text-muted-foreground">
        ({rating.count}条点评)
      </span>
    </div>
  )
}
