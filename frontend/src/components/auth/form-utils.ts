export function fieldError(errors: unknown[]) {
  return errors.length > 0 ? String(errors[0]) : null
}
