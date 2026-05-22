import { RiStarFill, RiStarLine } from "@remixicon/react"
import { cn } from "@/lib/utils"

interface RatingStarsProps {
  value: number
  onChange?: (value: number) => void
  size?: "sm" | "md" | "lg"
  readOnly?: boolean
}

export function RatingStars({
  value,
  onChange,
  size = "md",
  readOnly,
}: RatingStarsProps) {
  const sizeClass = size === "sm" ? "size-4" : size === "lg" ? "size-6" : "size-5"
  const interactive = !readOnly && onChange

  return (
    <div className="inline-flex items-center gap-0.5">
      {[1, 2, 3, 4, 5].map((star) => {
        const filled = star <= value
        const Icon = filled ? RiStarFill : RiStarLine
        return (
          <button
            key={star}
            type="button"
            disabled={!interactive}
            onClick={() => onChange?.(star)}
            className={cn(
              "p-0.5 transition-colors",
              interactive && "cursor-pointer hover:scale-110",
              !interactive && "cursor-default"
            )}
            aria-label={`${star} star`}
          >
            <Icon
              className={cn(
                sizeClass,
                filled ? "text-yellow-400" : "text-muted-foreground/40"
              )}
            />
          </button>
        )
      })}
    </div>
  )
}
