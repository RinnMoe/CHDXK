import { readFile } from "node:fs/promises"
import { join } from "node:path"

import { describe, expect, it } from "vitest"

import { BkjwClient, choosePageSize, searchUrl } from "./client.js"

describe("bkjw client helpers", () => {
  it("uses the largest page size exposed by the site without a hardcoded ceiling", () => {
    expect(choosePageSize(5000, [20, 2000, 1000])).toBe(2000)
    expect(choosePageSize(700, [20, 2000, 1000])).toBe(20)
  })

  it("builds the semester search URL with encoded query values", () => {
    const url = new URL(`http://bkjw.chd.edu.cn${searchUrl("term/1", 2, 1000)}`)
    expect(url.pathname).toBe("/eams/stdSyllabus!search.action")
    expect(url.searchParams.get("lesson.project.id")).toBe("1")
    expect(url.searchParams.get("lesson.semester.id")).toBe("term/1")
    expect(url.searchParams.get("pageNo")).toBe("2")
    expect(url.searchParams.get("pageSize")).toBe("1000")
  })

  it("fetches detail HTML through the authenticated request context", async () => {
    const html = await readFile(join(process.cwd(), "test", "fixtures", "detail.html"), "utf8")
    const requests: Array<{ url: string; options: Record<string, unknown> }> = []
    const context = {
      request: {
        get: async (url: string, options: Record<string, unknown>) => {
          requests.push({ url, options })
          return {
            status: () => 200,
            url: () => url,
            headers: () => ({}),
            text: async () => html,
          }
        },
      },
    }
    const detail = await new BkjwClient(context as never).fetchDetail(
      "/eams/stdSyllabus!info.action?lesson.id=101",
    )

    expect(detail.basic.courseCode).toBe("CS101")
    expect(requests).toHaveLength(1)
    expect(requests[0]).toMatchObject({
      url: "http://bkjw.chd.edu.cn/eams/stdSyllabus!info.action?lesson.id=101",
      options: { failOnStatusCode: false, maxRedirects: 0, timeout: 30_000 },
    })
  })

  it("promotes only CHD session cookies before a persistent context closes", async () => {
    const added: Array<Array<{ name: string; domain: string; expires: number }>> = []
    const context = {
      cookies: async () => [
        {
          name: "auth",
          value: "opaque",
          domain: "bkjw.chd.edu.cn",
          path: "/",
          expires: -1,
          httpOnly: true,
          secure: false,
          sameSite: "Lax" as const,
        },
        {
          name: "unrelated",
          value: "opaque",
          domain: "example.test",
          path: "/",
          expires: -1,
          httpOnly: false,
          secure: false,
          sameSite: "Lax" as const,
        },
      ],
      addCookies: async (cookies: Array<{ name: string; domain: string; expires: number }>) => {
        added.push(cookies)
      },
    }
    const client = new BkjwClient(context as never)

    await client.persistSessionCookies()

    expect(added).toHaveLength(1)
    expect(added[0]).toHaveLength(1)
    expect(added[0]?.[0]).toMatchObject({ name: "auth", domain: "bkjw.chd.edu.cn" })
    expect(added[0]?.[0]?.expires).toBeGreaterThan(Math.floor(Date.now() / 1000))
  })
})
