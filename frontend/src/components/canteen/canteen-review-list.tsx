import { Link } from "@tanstack/react-router"

import type { CanteenReviewDTO } from "@/api/canteen"
import { CanteenEmptyState } from "@/components/canteen/canteen-empty-state"
import { CanteenVoteButtons } from "@/components/canteen/canteen-vote-buttons"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ""
  return date.toLocaleDateString("zh-CN")
}

export function CanteenReviewList({
  reviews,
  currentUserID,
  showStore = false,
  emptyTitle = "还没有到店评价",
  emptyDescription = "发布第一条评价，帮助同学了解这家店。",
  onVote,
  onReport,
  voting,
}: {
  reviews: CanteenReviewDTO[]
  currentUserID?: number
  showStore?: boolean
  emptyTitle?: string
  emptyDescription?: string
  onVote?: (reviewID: number, voteType: number) => void
  onReport?: (reviewID: number) => void
  voting?: boolean
}) {
  if (reviews.length === 0) {
    return (
      <CanteenEmptyState title={emptyTitle} description={emptyDescription} />
    )
  }
  return (
    <div className="space-y-3">
      {reviews.map((review) => {
        const own = currentUserID !== undefined && review.user_id === currentUserID
        return (
          <Card key={review.id} size="sm">
            <CardHeader>
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 space-y-1">
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
                    {review.status !== "active" ? " · 已隐藏" : ""}
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
            <CardContent className="space-y-3">
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
              <div className="flex flex-wrap items-center justify-between gap-3">
                <p className="text-xs text-muted-foreground">
                  味道 {review.taste} · 性价比 {review.value} · 排队{" "}
                  {review.queue} · 卫生 {review.hygiene} · 服务 {review.service}
                </p>
                {onVote && (
                  <CanteenVoteButtons
                    likeCount={review.like_count}
                    dislikeCount={review.dislike_count}
                    myVote={review.my_vote}
                    disabled={own || voting}
                    onVote={(voteType) => onVote(review.id, voteType)}
                  />
                )}
                {onReport && !own && (
                  <button
                    type="button"
                    className="text-xs text-muted-foreground hover:underline"
                    onClick={() => onReport(review.id)}
                  >
                    举报
                  </button>
                )}
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
