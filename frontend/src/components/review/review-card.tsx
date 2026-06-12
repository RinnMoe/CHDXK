import { useEffect, useRef, useState } from "react"
import { Link } from "@tanstack/react-router"
import { RiEditLine, RiShareLine, RiWrenchLine } from "@remixicon/react"
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
import { VoteButtons } from "./vote-buttons"
import { ReviewContent } from "./review-content"
import { ReviewCardTime } from "./review-card-time"
import {
  ModeratorRemarkBanner,
  ModeratorRemarkDialog,
} from "./moderator-remark-dialog"
import type { ReviewDTO } from "@/api/review"

interface ReviewCardProps {
  review: ReviewDTO
  showCourse?: boolean
  showVoteButtons?: boolean
}

function getReviewUrl(reviewID: number) {
  const path = `/review/${reviewID}`
  if (typeof window === "undefined") return path
  return new URL(path, window.location.origin).toString()
}

async function copyText(text: string) {
  if (!navigator.clipboard?.writeText || !window.isSecureContext) {
    throw new Error("Clipboard API is unavailable")
  }

  await navigator.clipboard.writeText(text)
}

export function ReviewCard({
  review,
  showCourse = false,
  showVoteButtons = true,
}: ReviewCardProps) {
  const { user } = useAuth()
  const { mutateAsync: deleteReview } = useDeleteReview()
  const [copied, setCopied] = useState(false)
  const [remarkDialogOpen, setRemarkDialogOpen] = useState(false)
  const [manageMenuOpen, setManageMenuOpen] = useState(false)
  const [manageMenuSide, setManageMenuSide] = useState<"top" | "bottom">(
    "bottom"
  )
  const manageMenuTriggerRef = useRef<HTMLButtonElement | null>(null)
  const resetCopiedTimer = useRef<number | undefined>(undefined)
  const isAdmin = user?.is_admin() ?? false
  const canEdit =
    review.user_id != null &&
    user != null &&
    (user.id === review.user_id || isAdmin)
  const canManage = isAdmin

  useEffect(() => {
    return () => window.clearTimeout(resetCopiedTimer.current)
  }, [])

  async function handleShare() {
    try {
      await copyText(getReviewUrl(review.id))
      setCopied(true)
      window.clearTimeout(resetCopiedTimer.current)
      resetCopiedTimer.current = window.setTimeout(() => setCopied(false), 1600)
    } catch {
      setCopied(false)
    }
  }

  async function handleDelete() {
    await deleteReview(review.id)
  }

  function handleManageMenuOpenChange(open: boolean) {
    if (open && manageMenuTriggerRef.current) {
      const rect = manageMenuTriggerRef.current.getBoundingClientRect()
      setManageMenuSide(rect.top < window.innerHeight / 2 ? "bottom" : "top")
    }
    setManageMenuOpen(open)
  }

  return (
    <article className="border-b px-2 py-2 transition-colors hover:bg-muted/30">
      <div className="space-y-2">
        {showCourse && review.course && (
          <Link
            to="/course/$courseID"
            params={{ courseID: String(review.course.id) }}
            className="flex items-center gap-2 rounded-md transition-colors hover:bg-muted/50"
          >
            <span className="font-mono text-sm text-muted-foreground">
              {review.course.code}
            </span>
            <span className="font-semibold">{review.course.name}</span>
            <span
              className={`text-sm text-muted-foreground ${
                review.course.main_teacher.name.length >= 3 ? "min-w-[3em]" : ""
              }`}
            >
              {review.course.main_teacher.name}
            </span>
          </Link>
        )}

        <div className="flex items-center gap-2">
          <Link
            to="/review/$reviewID"
            params={{ reviewID: String(review.id) }}
            className="font-mono text-sm text-muted-foreground hover:text-foreground"
          >
            #{review.id}
          </Link>
          <ReviewCardTime review={review} />
        </div>

        <ModeratorRemarkBanner remark={review.moderator_remark} />

        <ReviewContent
          rating={review.rating}
          semester={review.semester}
          score={review.score}
          content={review.content}
        />

        <div className="flex items-center justify-start gap-1">
          {showVoteButtons && (
            <VoteButtons
              reviewID={review.id}
              likeCount={review.vote.like_count}
              dislikeCount={review.vote.dislike_count}
              myVote={review.vote.my_vote}
            />
          )}

          <div className="flex items-center gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleShare}
              className="gap-1 hover:text-inherit"
              aria-label="复制点评链接"
              title="复制点评链接"
            >
              <RiShareLine data-icon="inline-start" />
              {copied && <span className="text-sm">已复制</span>}
            </Button>
            {canEdit && (
              <Button
                asChild
                variant="ghost"
                size="sm"
                className="size-8 hover:text-inherit"
                aria-label="修改点评"
                title="修改点评"
              >
                <Link
                  to="/review/$reviewID/edit"
                  params={{ reviewID: String(review.id) }}
                >
                  <RiEditLine className="size-4" />
                </Link>
              </Button>
            )}
            {canManage && (
              <DropdownMenu
                open={manageMenuOpen}
                onOpenChange={handleManageMenuOpenChange}
              >
                <DropdownMenuTrigger asChild>
                  <Button
                    ref={manageMenuTriggerRef}
                    variant="ghost"
                    size="sm"
                    className="size-8 hover:text-inherit"
                  >
                    <RiWrenchLine className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" side={manageMenuSide}>
                  {isAdmin && (
                    <DropdownMenuItem onClick={() => setRemarkDialogOpen(true)}>
                      {review.moderator_remark ? "修改批注" : "添加批注"}
                    </DropdownMenuItem>
                  )}
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
            {isAdmin && (
              <ModeratorRemarkDialog
                review={review}
                open={remarkDialogOpen}
                onOpenChange={setRemarkDialogOpen}
              />
            )}
          </div>
        </div>
      </div>
    </article>
  )
}
