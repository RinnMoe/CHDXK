import { RiThumbUpLine, RiThumbDownLine } from "@remixicon/react"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import type { ReviewDTO } from "@/api/course"

interface ReviewItemProps {
  review: ReviewDTO
}

function formatDate(iso: string) {
  const d = new Date(iso)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  return `${y}-${m}-${day}`
}

export function ReviewItem({ review }: ReviewItemProps) {
  return (
    <Card>
      <CardContent className="pt-4 space-y-3">
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{review.score}</Badge>
          <Badge variant="outline">{review.rating} 星</Badge>
          <span className="ml-auto text-xs text-muted-foreground">
            {formatDate(review.created_at)}
          </span>
        </div>
        <p className="text-sm whitespace-pre-wrap leading-relaxed">
          {review.content}
        </p>
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-1">
            <RiThumbUpLine className="size-3.5" />
            {review.vote.like_count}
          </span>
          <span className="inline-flex items-center gap-1">
            <RiThumbDownLine className="size-3.5" />
            {review.vote.dislike_count}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
