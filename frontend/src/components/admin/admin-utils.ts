export type FormSubmitEvent = {
  preventDefault(): void
  currentTarget: HTMLFormElement
}

export function getErrorMessage(error: unknown) {
  if (error instanceof Error) return error.message
  return "请求失败"
}
