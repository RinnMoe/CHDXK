import { copy } from "@/lib/copy"
import { useMemo, useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { useReviewRevisions } from "@/hooks/use-review"
import { formatDateTime, formatReviewCardTime } from "@/lib/date"
import { RiArrowLeftSLine, RiArrowRightSLine } from "@remixicon/react"
import { ReviewContent } from "./review-content"
import type { ReviewDTO } from "@/api/review"

const REVISION_PAGE_SIZE = 1

interface ReviewRevisionsDialogProps {
  review: ReviewDTO
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ReviewRevisionsDialog({
  review,
  open,
  onOpenChange,
}: ReviewRevisionsDialogProps) {
  const [page, setPage] = useState(1)
  const {
    data: revisions,
    isLoading,
    isError,
  } = useReviewRevisions(review.id, open)
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
    onOpenChange(nextOpen)
    if (!nextOpen) setPage(1)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[80vh] grid-rows-[auto_minmax(0,1fr)_auto] overflow-hidden sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {copy(
              "copy.user.components.review.review-revisions-dialog.b0b89549bc8b"
            )}
            {review.id}
            {copy(
              "copy.user.components.review.review-revisions-dialog.f79958532e61"
            )}
          </DialogTitle>
          <DialogDescription>
            {copy(
              "copy.user.components.review.review-revisions-dialog.b3fc4e89095e"
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="min-h-0 overflow-y-auto pr-1">
          {isLoading ? (
            <div className="py-10 text-center text-sm text-muted-foreground">
              {copy(
                "copy.user.components.review.review-revisions-dialog.9dc0825fba54"
              )}
            </div>
          ) : isError ? (
            <div className="py-10 text-center text-sm text-destructive">
              {copy(
                "copy.user.components.review.review-revisions-dialog.918eec070450"
              )}
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
              {copy(
                "copy.user.components.review.review-revisions-dialog.a9f13188a317"
              )}
            </div>
          )}
        </div>

        {revisions?.length && currentRevision ? (
          <div className="flex items-center justify-between gap-3 border-t pt-4">
            <span
              className="text-xs text-muted-foreground tabular-nums"
              title={formatDateTime(currentRevision.created_at)}
            >
              {formatReviewCardTime(currentRevision.created_at)}
            </span>
            <div className="flex items-center gap-3">
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={copy(
                  "copy.user.components.review.review-revisions-dialog.f1a0c64eb1e2"
                )}
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
                aria-label={copy(
                  "copy.user.components.review.review-revisions-dialog.97bbce3408a9"
                )}
                disabled={currentPage >= totalPages}
                onClick={() => setPage(currentPage + 1)}
              >
                <RiArrowRightSLine />
              </Button>
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
