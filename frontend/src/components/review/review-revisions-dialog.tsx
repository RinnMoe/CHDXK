import { useMemo, useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { useReviewRevisions } from "@/hooks/use-review"
import { formatDateTime } from "@/lib/date"
import { RiArrowLeftSLine, RiArrowRightSLine } from "@remixicon/react"
import { ReviewContent } from "./review-content"
import type { ReviewDTO } from "@/api/review"

const REVISION_PAGE_SIZE = 1

interface ReviewRevisionsDialogProps {
  review: ReviewDTO
}

function isEdited(review: ReviewDTO) {
  if (!review.updated_at || !review.created_at) return false
  return (
    new Date(review.updated_at).getTime() -
      new Date(review.created_at).getTime() >
    1000
  )
}

export function ReviewRevisionsDialog({ review }: ReviewRevisionsDialogProps) {
  const { user } = useAuth()
  const [open, setOpen] = useState(false)
  const [page, setPage] = useState(1)
  const edited = isEdited(review)
  const displayTime = edited ? review.updated_at : review.created_at
  const canViewRevisions = edited && user?.role === "admin"
  const {
    data: revisions,
    isLoading,
    isError,
  } = useReviewRevisions(review.id, open && canViewRevisions)
  const totalPages = Math.max(
    1,
    Math.ceil((revisions?.length ?? 0) / REVISION_PAGE_SIZE)
  )
  const currentPage = Math.min(page, totalPages)
  const pagedRevisions = useMemo(() => {
    const start = (currentPage - 1) * REVISION_PAGE_SIZE
    return (revisions ?? []).slice(start, start + REVISION_PAGE_SIZE)
  }, [currentPage, revisions])
  const currentRevision = pagedRevisions[0]

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen)
    if (!nextOpen) setPage(1)
  }

  if (!edited) {
    return (
      <span className="ml-auto text-sm text-muted-foreground tabular-nums">
        {formatDateTime(displayTime)}
      </span>
    )
  }

  const timeControl = canViewRevisions ? (
    <button
      type="button"
      className="ml-auto rounded-sm text-sm text-muted-foreground tabular-nums transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none"
      onClick={() => setOpen(true)}
    >
      {formatDateTime(displayTime)}
      <span className="ml-1 text-muted-foreground/70">(已修改)</span>
    </button>
  ) : (
    <span className="ml-auto text-sm text-muted-foreground tabular-nums">
      {formatDateTime(displayTime)}
      <span className="ml-1 text-muted-foreground/70">(已修改)</span>
    </span>
  )

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>{timeControl}</TooltipTrigger>
          <TooltipContent>
            创建于 {formatDateTime(review.created_at)}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      {canViewRevisions && (
        <DialogContent className="max-h-[80vh] grid-rows-[auto_minmax(0,1fr)_auto] overflow-hidden sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>点评 #{review.id} 的修订历史</DialogTitle>
          </DialogHeader>

          <div className="min-h-0 overflow-y-auto pr-1">
            {isLoading ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                加载中...
              </div>
            ) : isError ? (
              <div className="py-10 text-center text-sm text-destructive">
                加载 revision 失败
              </div>
            ) : revisions?.length ? (
              <div className="space-y-5">
                {pagedRevisions.map((revision) => {
                  return (
                    <section key={revision.id}>
                      <ReviewContent
                        rating={revision.rating}
                        semester={revision.semester}
                        score={revision.score}
                        content={revision.content}
                      />
                    </section>
                  )
                })}
              </div>
            ) : (
              <div className="py-10 text-center text-sm text-muted-foreground">
                暂无 revision 记录
              </div>
            )}
          </div>

          {revisions?.length && currentRevision ? (
            <div className="flex items-center justify-between gap-3 border-t pt-4">
              <span className="text-xs text-muted-foreground tabular-nums">
                {formatDateTime(currentRevision.created_at)}
              </span>
              <div className="flex items-center gap-3">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label="上一条 revision"
                  disabled={currentPage <= 1}
                  onClick={() => setPage(currentPage - 1)}
                >
                  <RiArrowLeftSLine />
                </Button>
                <span className="min-w-16 text-center text-sm font-medium text-muted-foreground tabular-nums">
                  {currentPage} / {totalPages}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label="下一条 revision"
                  disabled={currentPage >= totalPages}
                  onClick={() => setPage(currentPage + 1)}
                >
                  <RiArrowRightSLine />
                </Button>
              </div>
            </div>
          ) : null}
        </DialogContent>
      )}
    </Dialog>
  )
}
