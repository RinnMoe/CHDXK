import { useEffect, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { RiShareForwardLine } from "@remixicon/react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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

function getReviewUrl(reviewID: number) {
  const path = `/reviews/${reviewID}`
  if (typeof window === "undefined") return path
  return new URL(path, window.location.origin).toString()
}

async function copyText(text: string) {
  if (navigator.clipboard?.writeText && window.isSecureContext) {
    await navigator.clipboard.writeText(text)
    return
  }

  const textarea = document.createElement("textarea")
  textarea.value = text
  textarea.setAttribute("readonly", "")
  textarea.style.position = "fixed"
  textarea.style.top = "0"
  textarea.style.left = "0"
  textarea.style.opacity = "0"
  document.body.appendChild(textarea)
  textarea.select()

  try {
    const copied = document.execCommand("copy")
    if (!copied) throw new Error("Copy command failed")
  } finally {
    document.body.removeChild(textarea)
  }
}

export function ReviewCard({
  review,
  showCourse = false,
  showVoteButtons = true,
}: ReviewCardProps) {
  const [copied, setCopied] = useState(false)
  const resetCopiedTimer = useRef<number | undefined>(undefined)
  const edited = isEdited(review)
  const displayTime = edited ? review.updated_at : review.created_at
  const tooltip = edited
    ? `创建于 ${formatDateTime(review.created_at)}`
    : undefined

  useEffect(() => {
    return () => window.clearTimeout(resetCopiedTimer.current)
  }, [])

  async function handleShare() {
    await copyText(getReviewUrl(review.id))
    setCopied(true)
    window.clearTimeout(resetCopiedTimer.current)
    resetCopiedTimer.current = window.setTimeout(() => setCopied(false), 1600)
  }

  return (
    <article className="border-b px-4 py-3 transition-colors hover:bg-muted/30">
      <div className="space-y-2">
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

        <div className="flex items-center justify-between gap-2">
          {showVoteButtons ? (
            <VoteButtons
              reviewID={review.id}
              likeCount={review.vote.like_count}
              dislikeCount={review.vote.dislike_count}
              myVote={review.vote.my_vote}
            />
          ) : (
            <span />
          )}

          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleShare}
            className="gap-1 text-muted-foreground"
            aria-label="复制评价链接"
            title="复制评价链接"
          >
            <RiShareForwardLine data-icon="inline-start" />
            {copied && <span className="text-xs">已复制</span>}
          </Button>
        </div>
      </div>
    </article>
  )
}
