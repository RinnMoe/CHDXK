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
      setError(err instanceof Error ? err.message : "保存失败")
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <DialogHeader>
        <DialogTitle>管理员批注</DialogTitle>
        <DialogDescription>
          点评 #{review.id} 的批注会对所有用户可见。
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-2">
        <Label htmlFor={`moderator-remark-${review.id}`}>批注内容</Label>
        <Textarea
          id={`moderator-remark-${review.id}`}
          value={remark}
          onChange={(e) => setRemark(e.target.value)}
          rows={5}
          className="resize-y text-sm"
          placeholder="填写管理员批注，清空后保存可移除批注"
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
          取消
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending ? "保存中..." : "保存批注"}
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
