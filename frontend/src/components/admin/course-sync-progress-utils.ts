export interface SyncProgress {
  stage: "listing" | "details" | "validating" | "importing" | "completed"
  completed: number
  total?: number
  failed: number
  retries: number
  page?: number
  pages?: number
  updated_at: string
  cached?: number
  retry_at?: string
}

export function progressPercent(progress: SyncProgress): number | undefined {
  if (!["listing", "details"].includes(progress.stage) || !progress.total)
    return undefined
  return Math.max(
    0,
    Math.min(100, Math.floor((progress.completed / progress.total) * 100))
  )
}
