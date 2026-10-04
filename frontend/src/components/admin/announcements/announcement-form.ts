import { copy } from "@/lib/copy"
import dayjs from "dayjs"
import type {
  AnnouncementDTO,
  SaveAnnouncementCommand,
} from "@/api/announcement"

export type AnnouncementFormState = {
  title: string
  body: string
  priority: string
  show_start: string
  show_end: string
  link_url: string
  link_title: string
}

const emptyForm: AnnouncementFormState = {
  title: "",
  body: "",
  priority: "0",
  show_start: "",
  show_end: "",
  link_url: "",
  link_title: "",
}

function toDateTimeLocal(value: string) {
  const date = dayjs(value)
  return date.isValid() ? date.format("YYYY-MM-DDTHH:mm") : ""
}

export function toAnnouncementFormState(
  announcement: AnnouncementDTO | null
): AnnouncementFormState {
  if (!announcement) return emptyForm
  return {
    title: announcement.title,
    body: announcement.body,
    priority: String(announcement.priority),
    show_start: toDateTimeLocal(announcement.show_start),
    show_end: toDateTimeLocal(announcement.show_end),
    link_url: announcement.link_url ?? "",
    link_title: announcement.link_title ?? "",
  }
}

export function buildAnnouncementCommand(
  form: AnnouncementFormState
): SaveAnnouncementCommand {
  const priority = Number(form.priority)
  return {
    title: form.title.trim(),
    body: form.body.trim(),
    priority: Number.isFinite(priority) ? Math.trunc(priority) : 0,
    show_start: new Date(form.show_start).toISOString(),
    show_end: new Date(form.show_end).toISOString(),
    link_url: form.link_url.trim(),
    link_title: form.link_title.trim(),
  }
}

export function validateAnnouncementForm(form: AnnouncementFormState) {
  if (!form.title.trim() && !form.body.trim())
    return copy(
      "copy.admin.components.admin.announcements.announcement-form.3e722daada31"
    )
  if (!form.show_start)
    return copy(
      "copy.admin.components.admin.announcements.announcement-form.a52d61ec812a"
    )
  if (!form.show_end)
    return copy(
      "copy.admin.components.admin.announcements.announcement-form.20f02738a747"
    )
  const start = new Date(form.show_start)
  const end = new Date(form.show_end)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return copy(
      "copy.admin.components.admin.announcements.announcement-form.ed14081e864b"
    )
  }
  if (end <= start)
    return copy(
      "copy.admin.components.admin.announcements.announcement-form.b913e107ad8b"
    )
  const linkURL = form.link_url.trim()
  if (linkURL) {
    try {
      const parsed = new URL(linkURL)
      if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
        return copy(
          "copy.admin.components.admin.announcements.announcement-form.46fe57101a00"
        )
      }
    } catch {
      return copy(
        "copy.admin.components.admin.announcements.announcement-form.b138dfd3b014"
      )
    }
  }
  return ""
}
