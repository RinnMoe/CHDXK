import type { CanteenHallDTO } from "@/api/canteen"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function CanteenHallFilter({
  halls,
  hallID,
  onSelect,
}: {
  halls: CanteenHallDTO[]
  hallID?: number
  onSelect: (hall: CanteenHallDTO | null) => void
}) {
  if (halls.length === 0) return null
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button
        type="button"
        size="sm"
        variant={hallID ? "outline" : "default"}
        onClick={() => onSelect(null)}
      >
        全部食堂
      </Button>
      {halls.map((hall) => (
        <Button
          key={hall.id}
          type="button"
          size="sm"
          variant={hallID === hall.id ? "default" : "outline"}
          className={cn(hall.store_count === 0 && "opacity-70")}
          onClick={() => onSelect(hall)}
        >
          {hall.short_name || hall.name}
          <span className="ml-1 text-xs opacity-70">{hall.store_count}</span>
        </Button>
      ))}
    </div>
  )
}
