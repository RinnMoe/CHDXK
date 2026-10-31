export type CanteenReviewDraftValues = {
  visitedAt: string
  taste: number
  value: number
  queue: number
  hygiene: number
  service: number
  content: string
  tags: string[]
}

type DraftEnvelope = {
  version: 1
  updatedAt: number
  values: CanteenReviewDraftValues
}

// The key includes the local user and store so switching accounts can never
// leak a draft into another account's editor.
export function canteenDraftKey(
  userID: number,
  storeID: number,
  mode: "new" | `edit:${number}`
) {
  return `canteen:review-draft:v1:${userID}:${storeID}:${mode}`
}

export function loadCanteenDraft(key: string): CanteenReviewDraftValues | null {
  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<DraftEnvelope>
    if (parsed.version !== 1 || !parsed.values) return null
    return parsed.values
  } catch {
    return null
  }
}

export function saveCanteenDraft(
  key: string,
  values: CanteenReviewDraftValues
) {
  try {
    const envelope: DraftEnvelope = {
      version: 1,
      updatedAt: Date.now(),
      values,
    }
    window.localStorage.setItem(key, JSON.stringify(envelope))
  } catch {
    // Drafts are best effort; private mode may reject writes.
  }
}

export function clearCanteenDraft(key: string) {
  try {
    window.localStorage.removeItem(key)
  } catch {
    // Ignore storage failures.
  }
}

export const emptyCanteenDraft: CanteenReviewDraftValues = {
  visitedAt: "",
  taste: 4,
  value: 4,
  queue: 4,
  hygiene: 4,
  service: 4,
  content: "",
  tags: [],
}
