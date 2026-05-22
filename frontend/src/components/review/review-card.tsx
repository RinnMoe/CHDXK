import { Link } from "react-router-dom"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
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
    <Card className="py-3 gap-0">
      <CardContent className="space-y-2">
        {showCourse && review.course && (
          <Link
            to={`/courses/${review.course.id}`}
            className="flex items-center gap-2 hover:bg-muted/50 -m-2 px-2 py-1.5 rounded-md transition-colors text-sm mb-2"
          >
            <span className="font-mono text-xs text-muted-foreground">
              {review.course.code}
            </span>
            <span className="font-medium truncate max-w-[200px]">
              {review.course.name}
            </span>
            <span className="text-xs text-muted-foreground truncate max-w-[120px]">
              {review.course.main_teacher.name}
            </span>
          </Link>
        )}

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
          <VoteButtons
            reviewID={review.id}
            likeCount={review.vote.like_count}
            dislikeCount={review.vote.dislike_count}
            myVote={review.vote.my_vote}
          />
        )}
      </CardContent>
    </Card>
  )
}
