export type ReviewDraftValues = {
  rating: number
  semester: string
  score: string
  content: string
}

type ReviewDraft = {
  version: 1
  updatedAt: number
  values: ReviewDraftValues
}

function isReviewDraftValues(value: unknown): value is ReviewDraftValues {
  if (!value || typeof value !== "object") return false
  const draft = value as Partial<ReviewDraftValues>
  return (
    typeof draft.rating === "number" &&
    typeof draft.semester === "string" &&
    typeof draft.score === "string" &&
    typeof draft.content === "string"
  )
}

export function loadReviewDraft(key: string): ReviewDraft | null {
  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) return null

    const draft = JSON.parse(raw) as Partial<ReviewDraft>
    if (draft.version !== 1 || !isReviewDraftValues(draft.values)) return null
    return {
      version: 1,
      updatedAt: typeof draft.updatedAt === "number" ? draft.updatedAt : 0,
      values: draft.values,
    }
  } catch {
    return null
  }
}

export function saveReviewDraft(key: string, values: ReviewDraftValues) {
  const draft: ReviewDraft = {
    version: 1,
    updatedAt: Date.now(),
    values,
  }

  try {
    window.localStorage.setItem(key, JSON.stringify(draft))
    return draft.updatedAt
  } catch {
    return null
  }
}

export function removeReviewDraft(key: string) {
  try {
    window.localStorage.removeItem(key)
  } catch {
    // ignore storage failures
  }
}
