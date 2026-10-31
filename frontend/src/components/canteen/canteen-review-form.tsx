import { useRef, useState, type ReactNode } from "react"

import type { CanteenReviewDraftValues } from "@/lib/canteen-draft"
import {
  clearCanteenDraft,
  emptyCanteenDraft,
  loadCanteenDraft,
  saveCanteenDraft,
} from "@/lib/canteen-draft"
import { CanteenRatingInput } from "@/components/canteen/canteen-rating-input"
import { CanteenTagPicker } from "@/components/canteen/canteen-tag-picker"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"

const ratingFields = [
  { key: "taste", label: "味道" },
  { key: "value", label: "性价比" },
  { key: "queue", label: "排队效率" },
  { key: "hygiene", label: "卫生环境" },
  { key: "service", label: "服务态度" },
] as const

export function CanteenReviewForm({
  draftKey,
  initial,
  submitLabel,
  submitting,
  error,
  extra,
  onSubmit,
}: {
  draftKey: string
  initial?: CanteenReviewDraftValues
  submitLabel: string
  submitting: boolean
  error?: string | null
  extra?: ReactNode
  onSubmit: (values: CanteenReviewDraftValues, idempotencyKey: string) => void
}) {
  const [values, setValues] = useState<CanteenReviewDraftValues>(() => {
    return loadCanteenDraft(draftKey) ?? initial ?? emptyCanteenDraft
  })
  const idempotencyKey = useRef(crypto.randomUUID())

  function update(next: Partial<CanteenReviewDraftValues>) {
    setValues((current) => {
      const merged = { ...current, ...next }
      saveCanteenDraft(draftKey, merged)
      return merged
    })
  }

  function submit() {
    clearCanteenDraft(draftKey)
    onSubmit(values, idempotencyKey.current)
  }

  return (
    <Card size="sm">
      <CardHeader>
        <h2 className="text-base font-semibold">到店评分</h2>
        <p className="text-xs text-muted-foreground">
          五项评分必填，文字与标签可选。同一店铺可以多次记录到店体验。
        </p>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="space-y-3">
          {ratingFields.map((field) => (
            <CanteenRatingInput
              key={field.key}
              label={field.label}
              value={values[field.key]}
              onChange={(value) => update({ [field.key]: value })}
            />
          ))}
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium" htmlFor="canteen-review-content">
            文字评价（可选）
          </label>
          <Textarea
            id="canteen-review-content"
            rows={4}
            maxLength={2000}
            placeholder="说说这次的体验…"
            value={values.content}
            onChange={(event) => update({ content: event.target.value })}
          />
          <p className="text-right text-xs text-muted-foreground">
            {values.content.length}/2000
          </p>
        </div>

        <div className="space-y-2">
          <p className="text-sm font-medium">体验标签（最多 5 个）</p>
          <CanteenTagPicker
            selected={values.tags}
            onChange={(tags) => update({ tags })}
          />
        </div>

        {extra}

        {error && <p className="text-sm text-destructive">{error}</p>}

        <Button type="button" onClick={submit} disabled={submitting}>
          {submitting ? "提交中…" : submitLabel}
        </Button>
      </CardContent>
    </Card>
  )
}
