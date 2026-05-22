import { Link } from "react-router-dom"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { VoteButtons } from "./vote-buttons"
import { RatingStars } from "./rating-stars"
import type { ReviewDTO } from "@/api/review"

interface ReviewCardProps {
  review: ReviewDTO
  showCourse?: boolean
  showVoteButtons?: boolean
}

function formatDateTime(iso: string) {
  const d = new Date(iso)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  const hh = String(d.getHours()).padStart(2, "0")
  const mm = String(d.getMinutes()).padStart(2, "0")
  return `${y}-${m}-${day} ${hh}:${mm}`
}

function isEdited(review: ReviewDTO) {
  if (!review.updated_at || !review.created_at) return false
  return (
    new Date(review.updated_at).getTime() -
      new Date(review.created_at).getTime() >
    1000
  )
}

export function ReviewCard({
  review,
  showCourse = false,
  showVoteButtons = true,
}: ReviewCardProps) {
  const edited = isEdited(review)
  const displayTime = edited ? review.updated_at : review.created_at
  const tooltip = edited
    ? `创建于 ${formatDateTime(review.created_at)}`
    : undefined

  return (
    <Card>
      <CardContent className="pt-4 space-y-3">
        {showCourse && review.course && (
          <Link
            to={`/courses/${review.course.id}`}
            className="flex items-start justify-between gap-2 hover:bg-muted/50 -m-2 p-2 rounded-md transition-colors"
          >
            <div className="min-w-0">
              <div className="text-xs text-muted-foreground font-mono">
                {review.course.code}
              </div>
              <div className="font-medium truncate">{review.course.name}</div>
              <div className="text-xs text-muted-foreground mt-0.5">
                {review.course.main_teacher.name}
              </div>
            </div>
          </Link>
        )}

        {showCourse && review.course && <Separator />}

        <div className="flex items-center gap-2 flex-wrap">
          <RatingStars value={review.rating} readOnly size="sm" />
          {review.score && review.score !== "未公布" && (
            <Badge variant="secondary">{review.score}</Badge>
          )}
          {review.semester && (
            <Badge variant="outline" className="font-mono">
              {review.semester}
            </Badge>
          )}
          <Link
            to={`/reviews/${review.id}`}
            className="text-xs text-muted-foreground font-mono hover:text-foreground"
          >
            #{review.id}
          </Link>
          <span
            className="ml-auto text-xs text-muted-foreground tabular-nums"
            title={tooltip}
          >
            {formatDateTime(displayTime)}
            {edited && (
              <span className="ml-1 text-muted-foreground/70">(已修改)</span>
            )}
          </span>
        </div>

        <p className="text-sm whitespace-pre-wrap leading-relaxed">
          {review.content}
        </p>

        {showVoteButtons && (
          <>
            <Separator />
            <VoteButtons
              reviewID={review.id}
              likeCount={review.vote.like_count}
              dislikeCount={review.vote.dislike_count}
              myVote={review.vote.my_vote}
            />
          </>
        )}
      </CardContent>
    </Card>
  )
}
