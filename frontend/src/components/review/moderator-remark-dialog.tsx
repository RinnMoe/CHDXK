import { copy } from "@/lib/copy"
import { useState, type SyntheticEvent } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { useUpdateReviewModeratorRemark } from "@/hooks/use-review"
import type { ReviewDTO } from "@/api/review"

interface ModeratorRemarkDialogProps {
  review: ReviewDTO
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModeratorRemarkDialog({
  review,
  open,
  onOpenChange,
}: ModeratorRemarkDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <ModeratorRemarkForm
          key={`${review.id}-${open ? review.moderator_remark : "closed"}`}
          review={review}
          onOpenChange={onOpenChange}
        />
      </DialogContent>
    </Dialog>
  )
}

function ModeratorRemarkForm({
  review,
  onOpenChange,
}: Pick<ModeratorRemarkDialogProps, "review" | "onOpenChange">) {
  const [remark, setRemark] = useState(review.moderator_remark ?? "")
  const [error, setError] = useState<string | null>(null)
  const { mutateAsync, isPending } = useUpdateReviewModeratorRemark()

  async function handleSubmit(e: SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    try {
      await mutateAsync({
        reviewID: review.id,
        cmd: { moderator_remark: remark },
      })
      onOpenChange(false)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : copy(
              "copy.user.components.review.moderator-remark-dialog.6309a3bb5ba4"
            )
      )
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <DialogHeader>
        <DialogTitle>
          {copy(
            "copy.user.components.review.moderator-remark-dialog.fe97e431532f"
          )}
        </DialogTitle>
        <DialogDescription>
          {copy(
            "copy.user.components.review.moderator-remark-dialog.b0b89549bc8b"
          )}
          {review.id}
          {copy(
            "copy.user.components.review.moderator-remark-dialog.8680c82f8d97"
          )}
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-2">
        <Label htmlFor={`moderator-remark-${review.id}`}>
          {copy(
            "copy.user.components.review.moderator-remark-dialog.6d5aba564779"
          )}
        </Label>
        <Textarea
          id={`moderator-remark-${review.id}`}
          value={remark}
          onChange={(e) => setRemark(e.target.value)}
          rows={5}
          className="resize-y text-sm"
          placeholder={copy(
            "copy.user.components.review.moderator-remark-dialog.bf6780b11d5f"
          )}
        />
      </div>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      <DialogFooter>
        <Button
          type="button"
          variant="ghost"
          onClick={() => onOpenChange(false)}
        >
          {copy(
            "copy.user.components.review.moderator-remark-dialog.2cd0f3be8738"
          )}
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending
            ? copy(
                "copy.user.components.review.moderator-remark-dialog.917b1c1f18d0"
              )
            : copy(
                "copy.user.components.review.moderator-remark-dialog.49dd2bd35dd7"
              )}
        </Button>
      </DialogFooter>
    </form>
  )
}

export function ModeratorRemarkBanner({ remark }: { remark: string }) {
  const trimmed = remark.trim()
  if (!trimmed) return null

  return (
    <div className="rounded-md border border-primary/20 bg-primary/5 px-3 py-2 text-sm">
      <div className="whitespace-pre-wrap text-foreground/90">{trimmed}</div>
    </div>
  )
}
