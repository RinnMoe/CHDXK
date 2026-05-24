import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function displayTeacherTitle(title?: string | null) {
  const normalizedTitle = title?.trim()
  return normalizedTitle && normalizedTitle !== "无" ? normalizedTitle : null
}
