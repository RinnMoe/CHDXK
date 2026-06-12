import {
  useCallback,
  useMemo,
  useRef,
  useState,
  type SyntheticEvent,
} from "react"
import { useForm } from "@tanstack/react-form"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  loadReviewDraft,
  removeReviewDraft,
  saveReviewDraft,
  type ReviewDraftValues,
} from "@/lib/review-draft"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { RatingStars } from "./rating-stars"
import { ReviewFormContentEditor } from "./review-form-content-editor"
import {
  CONTENT_MAX_LENGTH,
  CONTENT_MIN_LENGTH,
  DEFAULT_REVIEW_TEMPLATE,
  SCORE_MAX_LENGTH,
} from "./review-form-template"
import type {
  CreateReviewCommand,
  UpdateReviewCommand,
  ReviewDTO,
} from "@/api/review"

interface ReviewFormProps {
  courseID?: number
  initialReview?: ReviewDTO
  semesters?: string[]
  defaultSemester?: string
  onSubmit: (
    cmd: CreateReviewCommand | UpdateReviewCommand
  ) => Promise<void> | void
  onCancel?: () => void
  isSubmitting?: boolean
  draftUserID?: number
}

type ReviewFormValues = {
  rating: number
  semester: string
  score: string
  content: string
}

function fieldError(errors: unknown[]) {
  return errors.length > 0 ? String(errors[0]) : null
}

function reviewValuesEqual(a: ReviewFormValues, b: ReviewFormValues) {
  return (
    a.rating === b.rating &&
    a.semester === b.semester &&
    a.score === b.score &&
    a.content === b.content
  )
}

export function ReviewForm({
  courseID,
  initialReview,
  semesters,
  defaultSemester,
  onSubmit,
  onCancel,
  isSubmitting,
  draftUserID,
}: ReviewFormProps) {
  const isEdit = !!initialReview
  const availableSemesters = semesters ?? []
  const resolvedDefaultSemester =
    initialReview?.semester ?? defaultSemester ?? ""
  const [submitError, setSubmitError] = useState<string | null>(null)
  const draftKey = useMemo(() => {
    const owner = draftUserID ? `user:${draftUserID}` : "anonymous"
    if (initialReview)
      return `jcourse:review-draft:${owner}:edit:${initialReview.id}`
    if (courseID) return `jcourse:review-draft:${owner}:create:${courseID}`
    return null
  }, [courseID, draftUserID, initialReview])
  const baseValues = useMemo(
    () =>
      ({
        rating: initialReview?.rating ?? 0,
        semester: initialReview?.semester ?? "",
        score: initialReview?.score ?? "",
        content: initialReview?.content ?? DEFAULT_REVIEW_TEMPLATE,
      }) as ReviewFormValues,
    [initialReview]
  )
  const restoredDraft = useMemo(
    () => (draftKey ? loadReviewDraft(draftKey) : null),
    [draftKey]
  )
  const initialValues = useMemo(
    () => restoredDraft?.values ?? baseValues,
    [baseValues, restoredDraft]
  )
  const draftValuesRef = useRef<ReviewFormValues>(initialValues)
  const [draftSavedAt, setDraftSavedAt] = useState<number | null>(
    restoredDraft?.updatedAt ?? null
  )
  const [draftRestored, setDraftRestored] = useState(!!restoredDraft)

  function clearDraftState() {
    if (draftKey) removeReviewDraft(draftKey)
    setDraftSavedAt(null)
    setDraftRestored(false)
  }

  const updateDraft = useCallback(
    (patch: Partial<ReviewDraftValues>) => {
      const next = { ...draftValuesRef.current, ...patch }
      draftValuesRef.current = next
      if (!draftKey) return

      if (reviewValuesEqual(next, baseValues)) {
        removeReviewDraft(draftKey)
        setDraftSavedAt(null)
        setDraftRestored(false)
        return
      }

      const savedAt = saveReviewDraft(draftKey, next)
      if (savedAt) {
        setDraftSavedAt(savedAt)
        setDraftRestored(false)
      }
    },
    [baseValues, draftKey]
  )

  const form = useForm({
    defaultValues: initialValues,
    onSubmit: async ({ value }) => {
      const selectedSemester = getSelectedSemester(value.semester)
      if (!selectedSemester) return

      let cmd: CreateReviewCommand | UpdateReviewCommand
      if (isEdit) {
        cmd = {
          semester: selectedSemester,
          rating: value.rating,
          content: value.content,
          score: value.score || undefined,
        } as UpdateReviewCommand
      } else {
        if (!courseID) throw new Error("missing courseID")
        cmd = {
          course_id: courseID,
          semester: selectedSemester,
          rating: value.rating,
          content: value.content,
          score: value.score || undefined,
        } as CreateReviewCommand
      }

      await onSubmit(cmd)
      clearDraftState()
    },
  })

  function handleClearDraft() {
    clearDraftState()
    draftValuesRef.current = baseValues
    form.reset(baseValues)
  }

  function getSelectedSemester(semester: string) {
    const selectedSemester = semester || resolvedDefaultSemester
    return semesters &&
      selectedSemester &&
      !semesters.includes(selectedSemester)
      ? ""
      : selectedSemester
  }

  function handleSubmit(e: SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    e.stopPropagation()
    setSubmitError(null)
    void form.handleSubmit().catch((err: unknown) => {
      setSubmitError(err instanceof Error ? err.message : "提交失败")
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <form.Field
        name="rating"
        validators={{
          onSubmit: ({ value }) =>
            value < 1 || value > 5 ? "请选择评分（1-5 星）" : undefined,
        }}
      >
        {(field) => {
          const error = fieldError(field.state.meta.errors)
          return (
            <div className="space-y-2">
              <Label>评分</Label>
              <div>
                <RatingStars
                  value={field.state.value}
                  onChange={(value) => {
                    field.handleChange(value)
                    updateDraft({ rating: value })
                  }}
                  size="lg"
                />
              </div>
              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}
            </div>
          )
        }}
      </form.Field>

      <div className="grid gap-4 sm:grid-cols-2">
        <form.Field
          name="semester"
          validators={{
            onSubmit: ({ value }) =>
              getSelectedSemester(value) ? undefined : "请选择学期",
          }}
        >
          {(field) => {
            const selectedSemester = getSelectedSemester(field.state.value)
            const error = fieldError(field.state.meta.errors)
            return (
              <div className="space-y-2">
                <Label>学期</Label>
                <Select
                  value={selectedSemester}
                  onValueChange={(value) => {
                    field.handleChange(value)
                    updateDraft({ semester: value })
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="选择学期" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableSemesters.map((s) => (
                      <SelectItem key={s} value={s}>
                        {s}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {error && (
                  <p className="text-sm text-destructive" role="alert">
                    {error}
                  </p>
                )}
                <p className="text-sm leading-6 text-muted-foreground">
                  2026-2027 代表 2026-2027
                  学年度（2026.9-2027.8）。1代表秋季学期，2代表春季学期，3代表夏季学期/小学期。
                </p>
              </div>
            )
          }}
        </form.Field>
        <form.Field
          name="score"
          validators={{
            onSubmit: ({ value }) =>
              value.length > SCORE_MAX_LENGTH
                ? `分数最多 ${SCORE_MAX_LENGTH} 个字符`
                : undefined,
          }}
        >
          {(field) => {
            const error = fieldError(field.state.meta.errors)
            return (
              <div className="space-y-2">
                <Label htmlFor="score">分数（可选）</Label>
                <Input
                  id="score"
                  placeholder="如 A、92、中期退课（W）"
                  value={field.state.value}
                  onChange={(e) => {
                    field.handleChange(e.target.value)
                    updateDraft({ score: e.target.value })
                  }}
                  onBlur={field.handleBlur}
                  maxLength={SCORE_MAX_LENGTH}
                />
                {error && (
                  <p className="text-sm text-destructive" role="alert">
                    {error}
                  </p>
                )}
              </div>
            )
          }}
        </form.Field>
      </div>

      <form.Field
        name="content"
        validators={{
          onSubmit: ({ value }) => {
            if (!isEdit && value.trim() === DEFAULT_REVIEW_TEMPLATE.trim()) {
              return "请修改点评模板后再提交"
            }
            if (value.trim().length < CONTENT_MIN_LENGTH) {
              return `点评内容至少需要 ${CONTENT_MIN_LENGTH} 个字符`
            }
            if (value.length > CONTENT_MAX_LENGTH) {
              return `点评内容最多 ${CONTENT_MAX_LENGTH} 个字符`
            }
            return undefined
          },
        }}
      >
        {(field) => {
          const error = fieldError(field.state.meta.errors)
          const content = field.state.value
          function updateContent(nextContent: string) {
            field.handleChange(nextContent)
            updateDraft({ content: nextContent })
          }

          return (
            <ReviewFormContentEditor
              content={content}
              error={error}
              onChange={updateContent}
              onBlur={field.handleBlur}
            />
          )
        }}
      </form.Field>

      {submitError && (
        <p className="text-sm text-destructive" role="alert">
          {submitError}
        </p>
      )}

      {draftKey && (draftRestored || draftSavedAt) && (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
          <span>{draftRestored ? "已恢复本地草稿" : "本地草稿已保存"}</span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleClearDraft}
          >
            清除草稿
          </Button>
        </div>
      )}

      <div className="flex justify-end gap-2">
        {onCancel && (
          <Button type="button" variant="ghost" onClick={onCancel}>
            取消
          </Button>
        )}
        <form.Subscribe selector={(state) => state.isSubmitting}>
          {(formSubmitting) => {
            const submitting = isSubmitting || formSubmitting
            return (
              <Button type="submit" disabled={submitting}>
                {submitting ? "提交中..." : isEdit ? "更新点评" : "发布点评"}
              </Button>
            )
          }}
        </form.Subscribe>
      </div>
    </form>
  )
}
