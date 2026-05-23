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
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuth } from "@/contexts/auth-context"
import { useDeleteReview } from "@/hooks/use-review"
import { cn } from "@/lib/utils"
import { VoteButtons } from "./vote-buttons"
import { ReviewContent } from "./review-content"
import { ReviewRevisionsDialog } from "./review-revisions-dialog"
import type { ReviewDTO } from "@/api/review"

interface ReviewCardProps {
  review: ReviewDTO
  showCourse?: boolean
  showVoteButtons?: boolean
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
  const isOwnReview = review.user_id != null && user?.id === review.user_id
  const canManage =
    review.user_id != null &&
    user != null &&
    (user.id === review.user_id || user.role === "admin")

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
    <article
      className={cn(
        "border-b px-4 py-3 transition-colors",
        isOwnReview ? "bg-primary/5 hover:bg-primary/10" : "hover:bg-muted/30"
      )}
    >
      <div className="space-y-2">
        {showCourse && review.course && (
          <Link
            to={`/courses/${review.course.id}`}
            className="-m-2 mb-2 flex items-center gap-2 rounded-md px-2 py-1.5 transition-colors hover:bg-muted/50"
          >
            <span className="font-mono text-sm text-muted-foreground">
              {review.course.code}
            </span>
            <span className="max-w-[200px] truncate font-semibold">
              {review.course.name}
            </span>
            <span className="max-w-[120px] truncate text-sm text-muted-foreground">
              {review.course.main_teacher.name}
            </span>
          </Link>
        )}

        <div className="flex items-center gap-2">
          <Link
            to={`/reviews/${review.id}`}
            className="font-mono text-sm text-muted-foreground hover:text-foreground"
          >
            #{review.id}
          </Link>
          <ReviewRevisionsDialog review={review} />
        </div>

        <ReviewContent
          rating={review.rating}
          semester={review.semester}
          score={review.score}
          content={review.content}
        />

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
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-8 text-muted-foreground"
                  >
                    <RiWrenchLine className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem
                    onClick={() => navigate(`/reviews/${review.id}/edit`)}
                  >
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
                          className="text-destructive-foreground bg-destructive hover:bg-destructive/90"
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
