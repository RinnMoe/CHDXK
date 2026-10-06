import { renderToStaticMarkup } from "react-dom/server"
import { describe, expect, it } from "vitest"
import { TeacherProfile } from "./teacher-profile"
import type { TeacherDTO } from "@/api/teacher"

const teacher: TeacherDTO = {
  id: 1,
  code: "internal",
  name: "张三",
  department: "理学院",
}
describe("teacher profile summaries", () => {
  it("retains missing placeholders on detail pages", () => {
    const html = renderToStaticMarkup(<TeacherProfile teacher={teacher} />)
    expect(html.match(/暂无数据/g)).toHaveLength(6)
    expect(html).not.toContain("internal")
  })
  it("shows contact and office but sends long professional details to the homepage", () => {
    const html = renderToStaticMarkup(
      <TeacherProfile
        teacher={{
          ...teacher,
          profile: {
            url: "https://js.chd.edu.cn/t/list.htm",
            name: "张三",
            department: "理学院",
            title: "教授",
            updated_at: "",
            fields: {
              电子邮箱: "teacher@example.edu",
              联系电话: "029-12345678",
              办公地址: "教学楼101",
              研究领域: "数学",
              个人简介: "简".repeat(500),
              论文: "不展示全文",
            },
          },
        }}
      />
    )
    expect(html).toContain("前往教师主页")
    expect(html).toContain("teacher@example.edu")
    expect(html).toContain("029-12345678")
    expect(html).toContain("教学楼101")
    expect(html).toContain("数学")
    expect(html).not.toContain("不展示全文")
    expect(html).not.toContain("简".repeat(301))
  })
})
