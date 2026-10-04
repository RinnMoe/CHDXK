import { copy } from "@/lib/copy"
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
import {
  fieldError,
  reviewValuesEqual,
  type ReviewFormValues,
} from "./review-form-values"
import { ReviewSubmitButton } from "./review-submit-button"
import { ReviewDraftNotice } from "./review-draft-notice"
import { RatingStars } from "./rating-stars"
import { ReviewFormContentEditor } from "./review-form-content-editor"
import {
  CONTENT_MAX_LENGTH,
  CONTENT_MIN_LENGTH,
  DEFAULT_REVIEW_TEMPLATE,
  hasActualReviewContent,
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
      setSubmitError(
        err instanceof Error
          ? err.message
          : copy("copy.user.components.review.review-form.1440c7e23865")
      )
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <form.Field
        name="rating"
        validators={{
          onSubmit: ({ value }) =>
            value < 1 || value > 5
              ? copy("copy.user.components.review.review-form.e3bb52242abf")
              : undefined,
        }}
      >
        {(field) => {
          const error = fieldError(field.state.meta.errors)
          return (
            <div className="space-y-2">
              <Label>
                {copy("copy.user.components.review.review-form.96fff2e26c8d")}
              </Label>
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
              getSelectedSemester(value)
                ? undefined
                : copy("copy.user.components.review.review-form.c42296934f43"),
          }}
        >
          {(field) => {
            const selectedSemester = getSelectedSemester(field.state.value)
            const error = fieldError(field.state.meta.errors)
            return (
              <div className="space-y-2">
                <Label>
                  {copy("copy.user.components.review.review-form.79a4ef019e9c")}
                </Label>
                <Select
                  value={selectedSemester}
                  onValueChange={(value) => {
                    field.handleChange(value)
                    updateDraft({ semester: value })
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue
                      placeholder={copy(
                        "copy.user.components.review.review-form.522bedf98774"
                      )}
                    />
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
                  {copy("copy.user.components.review.review-form.130cb02ba7b2")}
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
                ? copy("copy.user.components.review.review-form.a25faaaa6b25", {
                    v0: SCORE_MAX_LENGTH,
                  })
                : undefined,
          }}
        >
          {(field) => {
            const error = fieldError(field.state.meta.errors)
            return (
              <div className="space-y-2">
                <Label htmlFor="score">
                  {copy("copy.user.components.review.review-form.86d12509c85f")}
                </Label>
                <Input
                  id="score"
                  placeholder={copy(
                    "copy.user.components.review.review-form.dd6d1ad1590a"
                  )}
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
            if (!hasActualReviewContent(value)) {
              return copy(
                "copy.user.components.review.review-form.ad57c41ed428"
              )
            }
            if (value.trim().length < CONTENT_MIN_LENGTH) {
              return copy(
                "copy.user.components.review.review-form.82efb72bee87",
                { v0: CONTENT_MIN_LENGTH }
              )
            }
            if (value.length > CONTENT_MAX_LENGTH) {
              return copy(
                "copy.user.components.review.review-form.55d4ef829a9d",
                { v0: CONTENT_MAX_LENGTH }
              )
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

      <ReviewDraftNotice
        visible={!!draftKey && !!(draftRestored || draftSavedAt)}
        draftRestored={draftRestored}
        onClear={handleClearDraft}
      />

      <div className="flex justify-end gap-2">
        {onCancel && (
          <Button type="button" variant="ghost" onClick={onCancel}>
            {copy("copy.user.components.review.review-form.2cd0f3be8738")}
          </Button>
        )}
        <form.Subscribe selector={(state) => state.isSubmitting}>
          {(formSubmitting) => (
            <ReviewSubmitButton
              submitting={!!isSubmitting || formSubmitting}
              isEdit={isEdit}
            />
          )}
        </form.Subscribe>
      </div>
    </form>
  )
}
