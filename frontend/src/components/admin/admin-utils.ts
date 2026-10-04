import { copy } from "@/lib/copy"

export type FormSubmitEvent = {
  preventDefault(): void
  currentTarget: HTMLFormElement
}

export function getErrorMessage(error: unknown) {
  if (error instanceof Error) return error.message
  return copy("copy.admin.components.admin.admin-utils.126fbf59cf04")
}
