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
  if (!form.title.trim() && !form.body.trim()) return "标题和内容至少填写一项"
  if (!form.show_start) return "请选择开始时间"
  if (!form.show_end) return "请选择结束时间"
  const start = new Date(form.show_start)
  const end = new Date(form.show_end)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return "展示时间无效"
  }
  if (end <= start) return "结束时间必须晚于开始时间"
  const linkURL = form.link_url.trim()
  if (linkURL) {
    try {
      const parsed = new URL(linkURL)
      if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
        return "外链只支持 http 或 https"
      }
    } catch {
      return "外链地址无效"
    }
  }
  return ""
}
