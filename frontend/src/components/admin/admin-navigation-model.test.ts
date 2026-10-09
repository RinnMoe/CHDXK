import { describe, expect, it } from "vitest"
import { normalizeAuthUser } from "@/api/auth"
import { canAccessAdminTab, type AdminTab } from "./admin-navigation-model"

describe("课程数据与开发维护权限", () => {
  const sections = [
    { permission: "course_data", tabs: ["teacher-profiles", "course-data"] },
    { permission: "maintenance", tabs: ["system-api-key", "audit-log"] },
  ] as const
  const account = { id: 8, username: "admin", email: "", role: "admin" }

  for (const section of sections) {
    it(`${section.permission} 仅超管或获得下放的管理员可访问`, () => {
      for (const role of ["admin", "user", "super_admin"]) {
        for (const permissions of [[], [section.permission]]) {
          const user = normalizeAuthUser({ ...account, role, permissions })
          for (const tab of section.tabs) {
            expect(canAccessAdminTab(tab as AdminTab, user.can)).toBe(
              role === "super_admin" ||
                (role === "admin" && permissions.includes(section.permission))
            )
          }
        }
      }
    })
  }

  it("两类下放互不授予，内容运营和用户处理仍可用", () => {
    const user = normalizeAuthUser({ ...account, permissions: ["course_data"] })
    expect(canAccessAdminTab("audit-log", user.can)).toBe(false)
    expect(canAccessAdminTab("users", user.can)).toBe(true)
    expect(canAccessAdminTab("announcement", user.can)).toBe(true)
  })
})
