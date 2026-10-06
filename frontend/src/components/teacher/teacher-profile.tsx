import { RiExternalLinkLine } from "@remixicon/react"
import { buttonVariants } from "@/components/ui/button"
import type { TeacherDTO } from "@/api/teacher"

const fields = ["职称", "学位", "联系方式", "办公地点", "研究方向", "个人简介"]
export function TeacherProfile({ teacher }: { teacher: TeacherDTO }) {
  return (
    <section className="space-y-4 rounded-lg border p-4" aria-label="教师资料">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="text-lg font-semibold">教师资料</h2>
        {teacher.profile?.url && (
          <a
            className={buttonVariants({ variant: "outline", size: "sm" })}
            href={teacher.profile.url}
            target="_blank"
            rel="noreferrer"
          >
            前往教师主页
            <RiExternalLinkLine className="size-4" aria-hidden="true" />
          </a>
        )}
      </div>
      <dl className="grid gap-4 sm:grid-cols-2">
        {fields.map((label) => {
          const value =
            label === "职称"
              ? teacher.profile?.title
              : label === "联系方式"
                ? [
                    teacher.profile?.fields["电子邮箱"],
                    teacher.profile?.fields["联系电话"],
                    teacher.profile?.fields["联系方式"],
                  ]
                    .filter(Boolean)
                    .join("\n")
                : label === "办公地点"
                  ? teacher.profile?.fields["办公地点"] ||
                    teacher.profile?.fields["办公地址"]
                  : label === "研究方向"
                    ? teacher.profile?.fields["研究方向"] ||
                      teacher.profile?.fields["研究领域"]
                    : teacher.profile?.fields[label]
          const summary = value?.trim()
          return (
            <div
              key={label}
              className={
                label === "个人简介" || label === "研究方向"
                  ? "sm:col-span-2"
                  : ""
              }
            >
              <dt className="text-sm font-medium">{label}</dt>
              <dd className="mt-1 text-sm wrap-break-word whitespace-pre-wrap text-muted-foreground">
                {summary
                  ? summary.length > 300
                    ? `${summary.slice(0, 300)}…`
                    : summary
                  : "暂无数据"}
              </dd>
            </div>
          )
        })}
      </dl>
    </section>
  )
}
