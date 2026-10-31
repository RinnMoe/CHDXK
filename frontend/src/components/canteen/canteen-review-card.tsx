import { Link } from "@tanstack/react-router"

import type { CanteenReviewDTO } from "@/api/canteen"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ""
  return date.toLocaleDateString("zh-CN")
}

export function CanteenReviewCard({
  review,
  showStore = false,
}: {
  review: CanteenReviewDTO
  showStore?: boolean
}) {
  return (
    <Card size="sm">
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            {showStore && (
              <Link
                to="/canteen/stores/$storeID"
                params={{ storeID: String(review.store_id) }}
                className="text-sm font-medium hover:underline"
              >
                {review.store_name}
              </Link>
            )}
            <p className="text-xs text-muted-foreground">
              {formatDate(review.visited_at)} 到店
            </p>
          </div>
          <div className="text-right">
            <div className="text-base font-semibold">
              {review.overall_score.toFixed(1)}
            </div>
            <div className="text-xs text-muted-foreground">综合分</div>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-2">
        {review.content && (
          <p className="text-sm whitespace-pre-wrap">{review.content}</p>
        )}
        {review.tags.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {review.tags.map((tag) => (
              <Badge key={tag} variant="outline">
                {tag}
              </Badge>
            ))}
          </div>
        )}
        <p className="text-xs text-muted-foreground">
          味道 {review.taste} · 性价比 {review.value} · 排队 {review.queue} · 卫生{" "}
          {review.hygiene} · 服务 {review.service}
        </p>
      </CardContent>
    </Card>
  )
}
