import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { useCreateCanteenReport } from "@/hooks/use-canteen-reports"

const reportReasons = [
  { value: "spam", label: "垃圾广告" },
  { value: "abuse", label: "辱骂攻击" },
  { value: "false_info", label: "虚假信息" },
  { value: "privacy", label: "隐私泄露" },
  { value: "other", label: "其他" },
] as const

export function CanteenReportDialog({
  open,
  onOpenChange,
  subjectType,
  subjectID,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  subjectType: "store" | "review" | "dish_review" | "media"
  subjectID: number
}) {
  const [reason, setReason] = useState<string>("spam")
  const [detail, setDetail] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [done, setDone] = useState(false)
  const create = useCreateCanteenReport()

  async function submit() {
    setError(null)
    try {
      await create.mutateAsync({
        subject_type: subjectType,
        subject_id: subjectID,
        reason_code: reason,
        detail,
      })
      setDone(true)
      setDetail("")
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : "举报提交失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>举报内容</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <select
            className="h-9 w-full rounded-md border border-input bg-transparent px-2.5 text-sm"
            value={reason}
            onChange={(event) => setReason(event.target.value)}
          >
            {reportReasons.map((item) => (
              <option key={item.value} value={item.value}>
                {item.label}
              </option>
            ))}
          </select>
          <Textarea
            rows={3}
            maxLength={500}
            placeholder="补充说明（可选）"
            value={detail}
            onChange={(event) => setDetail(event.target.value)}
          />
          {error && <p className="text-sm text-destructive">{error}</p>}
          {done && <p className="text-sm text-muted-foreground">已提交举报</p>}
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            取消
          </Button>
          <Button type="button" disabled={create.isPending} onClick={() => void submit()}>
            {create.isPending ? "提交中…" : "提交举报"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
