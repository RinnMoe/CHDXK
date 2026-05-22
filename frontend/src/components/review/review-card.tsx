import { useEffect, useRef, useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { RiShareForwardLine, RiWrenchLine } from "@remixicon/react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import Markdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { useAuth } from "@/contexts/auth-context"
import { useDeleteReview } from "@/hooks/use-review"
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
  const { user } = useAuth()
  const navigate = useNavigate()
  const { mutateAsync: deleteReview } = useDeleteReview()
  const [copied, setCopied] = useState(false)
  const resetCopiedTimer = useRef<number | undefined>(undefined)
  const edited = isEdited(review)
  const displayTime = edited ? review.updated_at : review.created_at
  const tooltip = edited
    ? `创建于 ${formatDateTime(review.created_at)}`
    : undefined
  const canManage = review.user_id != null && user != null && (user.id === review.user_id || user.role === "admin")

  useEffect(() => {
    return () => window.clearTimeout(resetCopiedTimer.current)
  }, [])

  async function handleShare() {
    await copyText(getReviewUrl(review.id))
    setCopied(true)
    window.clearTimeout(resetCopiedTimer.current)
    resetCopiedTimer.current = window.setTimeout(() => setCopied(false), 1600)
  }

  async function handleDelete() {
    await deleteReview(review.id)
  }

  return (
    <article className="border-b px-4 py-3 transition-colors hover:bg-muted/30">
      <div className="space-y-2">
        {showCourse && review.course && (
          <Link
            to={`/courses/${review.course.id}`}
            className="flex items-center gap-2 hover:bg-muted/50 -m-2 px-2 py-1.5 rounded-md transition-colors mb-2"
          >
            <span className="font-mono text-sm text-muted-foreground">
              {review.course.code}
            </span>
            <span className="font-semibold truncate max-w-[200px]">
              {review.course.name}
            </span>
            <span className="text-sm text-muted-foreground truncate max-w-[120px]">
              {review.course.main_teacher.name}
            </span>
          </Link>
        )}

        <div className="flex items-center gap-2">
          <Link
            to={`/reviews/${review.id}`}
            className="text-sm text-muted-foreground font-mono hover:text-foreground"
          >
            #{review.id}
          </Link>
          <span
            className="ml-auto text-sm text-muted-foreground tabular-nums"
            title={tooltip}
          >
            {formatDateTime(displayTime)}
            {edited && (
              <span className="ml-1 text-muted-foreground/70">(已修改)</span>
            )}
          </span>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <RatingStars value={review.rating} readOnly size="sm" />
          {review.semester && (
            <Badge variant="outline" className="font-mono">
              {review.semester}
            </Badge>
          )}
          {review.score && (
            <Badge variant="secondary">{review.score}</Badge>
          )}
        </div>

        <div className="text-sm leading-relaxed prose prose-sm max-w-none dark:prose-invert">
          <Markdown remarkPlugins={[remarkGfm]}>{review.content}</Markdown>
        </div>

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

          <div className="flex items-center gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleShare}
              className="gap-1 text-muted-foreground"
              aria-label="复制点评链接"
              title="复制点评链接"
            >
              <RiShareForwardLine data-icon="inline-start" />
              {copied && <span className="text-sm">已复制</span>}
            </Button>
            {canManage && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="size-8 text-muted-foreground">
                    <RiWrenchLine className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => navigate(`/reviews/${review.id}/edit`)}>
                    修改点评
                  </DropdownMenuItem>
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onSelect={(e) => e.preventDefault()}
                      >
                        删除点评
                      </DropdownMenuItem>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>确认删除</AlertDialogTitle>
                        <AlertDialogDescription>
                          删除后无法恢复，确定要删除这条点评吗？
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                          className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                          onClick={handleDelete}
                        >
                          删除
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>
        </div>
      </div>
    </article>
  )
}
