import { describe, expect, it } from "vitest"
import { fillTeacherProfile, teacherProfileForm } from "./teacher-profile-form"

describe("teacher profile fill", () => {
  it("keeps local identity and missing values until explicit save", () => {
    const form = teacherProfileForm({
      id: 1,
      code: "x",
      name: "张三",
      department: "理学院",
      title: "讲师",
    })
    form.fields["办公地点"] = "现有办公室"
    const result = fillTeacherProfile(form, {
      name: "其他名字",
      department: "其他学院",
      title: "教授",
      url: "https://js.chd.edu.cn/a/list.htm",
      updated_at: "",
      fields: {
        电子邮箱: "office@example.edu",
        联系电话: "123",
        研究领域: "数学",
      },
    })
    expect(result.name).toBe("张三")
    expect(result.department).toBe("理学院")
    expect(result.fields["办公地点"]).toBe("现有办公室")
    expect(result.fields["研究方向"]).toBe("数学")
    expect(result.fields["联系方式"]).toBe("office@example.edu\n123")
    expect(result.title).toBe("教授")
    expect(form.title).toBe("讲师")
    expect(form.fields["研究方向"]).toBe("")
  })
})
