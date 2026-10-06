import type { TeacherDTO, TeacherProfileDTO } from "@/api/teacher"

export function profileFormFields(p?: TeacherProfileDTO) {
  return {
    学位: p?.fields["学位"] ?? "",
    联系方式: [
      p?.fields["电子邮箱"],
      p?.fields["联系电话"],
      p?.fields["联系方式"],
    ]
      .filter(Boolean)
      .join("\n"),
    办公地点: p?.fields["办公地点"] || p?.fields["办公地址"] || "",
    研究方向: p?.fields["研究方向"] || p?.fields["研究领域"] || "",
    个人简介: p?.fields["个人简介"] ?? "",
  } as Record<string, string>
}
export function teacherProfileForm(teacher: TeacherDTO) {
  return {
    name: teacher.name,
    department: teacher.department,
    title: teacher.title ?? "",
    url: teacher.profile?.url ?? "",
    fields: profileFormFields(teacher.profile),
  }
}
export function fillTeacherProfile(
  form: ReturnType<typeof teacherProfileForm>,
  p: TeacherProfileDTO
) {
  const fields = { ...form.fields }
  for (const [key, value] of Object.entries(profileFormFields(p))) {
    if (value.trim()) fields[key] = value
  }
  return {
    ...form,
    title: p.title.trim() ? p.title : form.title,
    url: p.url,
    fields,
  }
}
