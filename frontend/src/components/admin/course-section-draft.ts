import { copy } from "@/lib/copy"
import type { CourseSectionDTO } from "@/api/course"

export interface SectionDraft {
  source_lesson_id: string
  lesson_no: string
  course_name: string
  course_type: string
  credit: string
  department: string
  language: string
  campus: string
  assessment: string
  teach_class: string
  actual_count: string
  limit_count: string
  course_period: string
  total_hours: string
  weekly_hours: string
  week_range: string
  start_week: string
  end_week: string
  week_count: string
  remark: string
  syllabus_url: string
  teachers: string
}

export const emptySection: SectionDraft = {
  source_lesson_id: "",
  lesson_no: "",
  course_name: "",
  course_type: "",
  credit: "0",
  department: "",
  language: "",
  campus: "",
  assessment: "",
  teach_class: "",
  actual_count: "",
  limit_count: "",
  course_period: "",
  total_hours: "",
  weekly_hours: "",
  week_range: "",
  start_week: "",
  end_week: "",
  week_count: "",
  remark: "",
  syllabus_url: "",
  teachers: "",
}

export function draftFromSection(section?: CourseSectionDTO): SectionDraft {
  if (!section) return { ...emptySection }
  return {
    source_lesson_id: section.source_lesson_id ?? "",
    lesson_no: section.lesson_no ?? "",
    course_name: section.course_name ?? "",
    course_type: section.course_type ?? "",
    credit: String(section.credit ?? 0),
    department: section.department ?? "",
    language: section.language ?? "",
    campus: section.campus ?? "",
    assessment: section.assessment ?? "",
    teach_class: section.teach_class ?? "",
    actual_count:
      section.actual_count == null ? "" : String(section.actual_count),
    limit_count: section.limit_count == null ? "" : String(section.limit_count),
    course_period: section.course_period ?? "",
    total_hours: section.total_hours == null ? "" : String(section.total_hours),
    weekly_hours:
      section.weekly_hours == null ? "" : String(section.weekly_hours),
    week_range: section.week_range ?? "",
    start_week: section.start_week == null ? "" : String(section.start_week),
    end_week: section.end_week == null ? "" : String(section.end_week),
    week_count: section.week_count == null ? "" : String(section.week_count),
    remark: section.remark ?? "",
    syllabus_url: section.syllabus_url ?? "",
    teachers: (section.teachers ?? []).join("，"),
  }
}

function optionalNumber(value: string, field: string) {
  if (!value.trim()) return null
  const number = Number(value)
  if (!Number.isFinite(number))
    throw new Error(
      copy("copy.admin.components.admin.course-management-tab.4396291ef58e", {
        v0: field,
      })
    )
  return number
}

export function sectionCommand(section: SectionDraft) {
  const credit = optionalNumber(
    section.credit,
    copy("copy.admin.components.admin.course-management-tab.f9d072c4804c")
  )
  if (credit == null)
    throw new Error(
      copy("copy.admin.components.admin.course-management-tab.44c1e84ed465")
    )
  return {
    source_lesson_id: section.source_lesson_id.trim(),
    lesson_no: section.lesson_no.trim(),
    course_name: section.course_name.trim(),
    course_type: section.course_type.trim(),
    credit,
    department: section.department.trim(),
    language: section.language.trim(),
    campus: section.campus.trim(),
    assessment: section.assessment.trim(),
    teach_class: section.teach_class.trim(),
    actual_count: optionalNumber(
      section.actual_count,
      copy("copy.admin.components.admin.course-management-tab.770bd60c8b1a")
    ),
    limit_count: optionalNumber(
      section.limit_count,
      copy("copy.admin.components.admin.course-management-tab.afc44894f290")
    ),
    course_period: section.course_period.trim(),
    total_hours: optionalNumber(
      section.total_hours,
      copy("copy.admin.components.admin.course-management-tab.6b7e4d50690c")
    ),
    weekly_hours: optionalNumber(
      section.weekly_hours,
      copy("copy.admin.components.admin.course-management-tab.c82cc02ad5ff")
    ),
    week_range: section.week_range.trim(),
    start_week: optionalNumber(
      section.start_week,
      copy("copy.admin.components.admin.course-management-tab.9ab9ff8f4d36")
    ),
    end_week: optionalNumber(
      section.end_week,
      copy("copy.admin.components.admin.course-management-tab.fa2920ccd890")
    ),
    week_count: optionalNumber(
      section.week_count,
      copy("copy.admin.components.admin.course-management-tab.c9462526d533")
    ),
    remark: section.remark.trim(),
    syllabus_url: section.syllabus_url.trim(),
    teachers: section.teachers
      .split(/[，,;；]/)
      .map((name) => name.trim())
      .filter(Boolean),
  }
}

export const sectionFields: [keyof SectionDraft, string, "text" | "number"][] =
  [
    [
      "lesson_no",
      copy("copy.admin.components.admin.course-management-tab.a7a9051c47e9"),
      "text",
    ],
    [
      "source_lesson_id",
      copy("copy.admin.components.admin.course-management-tab.54ec48e9d278"),
      "text",
    ],
    [
      "course_name",
      copy("copy.admin.components.admin.course-management-tab.bd2d74b297d1"),
      "text",
    ],
    [
      "course_type",
      copy("copy.admin.components.admin.course-management-tab.66e8e3adf36b"),
      "text",
    ],
    [
      "credit",
      copy("copy.admin.components.admin.course-management-tab.f9d072c4804c"),
      "number",
    ],
    [
      "department",
      copy("copy.admin.components.admin.course-management-tab.69907ec332f4"),
      "text",
    ],
    [
      "language",
      copy("copy.admin.components.admin.course-management-tab.2cbd78206ec3"),
      "text",
    ],
    [
      "campus",
      copy("copy.admin.components.admin.course-management-tab.406c59bfe6b8"),
      "text",
    ],
    [
      "assessment",
      copy("copy.admin.components.admin.course-management-tab.ad07afbbae6d"),
      "text",
    ],
    [
      "teach_class",
      copy("copy.admin.components.admin.course-management-tab.4c9c646f81d1"),
      "text",
    ],
    [
      "actual_count",
      copy("copy.admin.components.admin.course-management-tab.770bd60c8b1a"),
      "number",
    ],
    [
      "limit_count",
      copy("copy.admin.components.admin.course-management-tab.afc44894f290"),
      "number",
    ],
    [
      "course_period",
      copy("copy.admin.components.admin.course-management-tab.9e92320aa542"),
      "text",
    ],
    [
      "total_hours",
      copy("copy.admin.components.admin.course-management-tab.6b7e4d50690c"),
      "number",
    ],
    [
      "weekly_hours",
      copy("copy.admin.components.admin.course-management-tab.c82cc02ad5ff"),
      "number",
    ],
    [
      "week_range",
      copy("copy.admin.components.admin.course-management-tab.5d668f3c9656"),
      "text",
    ],
    [
      "start_week",
      copy("copy.admin.components.admin.course-management-tab.9ab9ff8f4d36"),
      "number",
    ],
    [
      "end_week",
      copy("copy.admin.components.admin.course-management-tab.fa2920ccd890"),
      "number",
    ],
    [
      "week_count",
      copy("copy.admin.components.admin.course-management-tab.c9462526d533"),
      "number",
    ],
  ]
