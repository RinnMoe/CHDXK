export type ReviewFormValues = {
  rating: number
  semester: string
  score: string
  content: string
}

export function fieldError(errors: unknown[]) {
  return errors.length > 0 ? String(errors[0]) : null
}

export function reviewValuesEqual(a: ReviewFormValues, b: ReviewFormValues) {
  return (
    a.rating === b.rating &&
    a.semester === b.semester &&
    a.score === b.score &&
    a.content === b.content
  )
}
