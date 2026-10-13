import { renderToStaticMarkup } from "react-dom/server"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { ReactNode } from "react"
import { mockCourses } from "@/mocks/fixtures/courses"
import { mockReviews } from "@/mocks/fixtures/reviews"
import { HomePage } from "./home-page"

const state = vi.hoisted(() => ({ queries: {} as Record<string, unknown> }))
vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => vi.fn(),
  Link: ({
    to,
    params,
    search,
    children,
    ...props
  }: {
    to: string
    params?: Record<string, string>
    search?: Record<string, unknown>
    children: ReactNode
  }) => {
    const path = to.replace("$courseID", params?.courseID ?? "")
    const query = new URLSearchParams()
    for (const [key, value] of Object.entries(search ?? {})) {
      if (typeof value === "string" || typeof value === "number") {
        query.set(key, String(value))
      } else if (Array.isArray(value)) {
        for (const item of value) {
          if (typeof item === "string") query.append(key, item)
        }
      }
    }
    return (
      <a href={path + (query.size ? `?${query}` : "")} {...props}>
        {children}
      </a>
    )
  },
}))
vi.mock("@/contexts/auth-context", () => ({
  useAuth: () => ({ user: { id: 1 } }),
}))
vi.mock("@/components/layout/page-shell", () => ({
  PageShell: ({ children }: { children: ReactNode }) => <main>{children}</main>,
}))
vi.mock("@/components/common/page-title", () => ({ PageTitle: () => null }))
vi.mock("@/components/course/enrollment-import-dialog", () => ({
  EnrollmentImportDialog: () => <button>导入课程</button>,
}))
vi.mock("@/components/review/review-list", () => ({
  ReviewList: ({ reviews }: { reviews: { id: number }[] }) => (
    <div>
      {reviews.map((review) => (
        <article key={review.id}>review-{review.id}</article>
      ))}
    </div>
  ),
}))
vi.mock("@/hooks/use-course", () => ({
  useCourseFilters: () => state.queries.filters,
  useCourseDiscovery: () => state.queries.discovery,
  useCourseReviewPrompts: () => state.queries.prompts,
  useHotCourses: () => state.queries.hot,
  useFollowedCourses: () => state.queries.followed,
}))
vi.mock("@/hooks/use-review", () => ({
  useReviews: () => state.queries.reviews,
}))

const loaded = (data: unknown) => ({
  data,
  isPending: false,
  isError: false,
  refetch: vi.fn(),
})

beforeEach(() => {
  state.queries = {
    filters: loaded({ categories: [{ name: "公共基础", count: 20 }] }),
    discovery: loaded({
      semester: "2026-2027-1",
      items: mockCourses.slice(0, 24).map((course) => ({
        ...course,
        rating: { ...course.rating, count: 0, avg: 0 },
      })),
    }),
    prompts: loaded({ has_enrollments: false, items: [] }),
    hot: loaded({ items: [] }),
    followed: loaded({ items: [] }),
    reviews: loaded({ items: [] }),
  }
})

describe("home content during a sparse community", () => {
  it("shows real courses, search, categories and contribution paths without empty ranking panels", () => {
    const html = renderToStaticMarkup(<HomePage />)
    expect(html).toContain('role="search"')
    expect(html).toContain("发现课程")
    expect(html).not.toContain("最近收录学期")
    expect(html).not.toContain("来源学期")
    expect(html.match(/浏览全部/g)).toHaveLength(1)
    expect(html).toContain(mockCourses[0].name)
    expect(html).not.toContain(mockCourses[6].name)
    expect(html.match(/暂无点评/g)).toHaveLength(6)
    expect(html).toContain("换一批")
    expect(html).toContain("导入课程")
    expect(html).toContain("搜索课程")
    expect(html).toContain("categories=%E5%85%AC%E5%85%B1%E5%9F%BA%E7%A1%80")
    expect(html).not.toContain('aria-label="热门课程"')
    expect(html).not.toContain("暂无热门课程")
    expect(html).not.toContain("暂无关注课程")
  })

  it.each([1, 2, 6, 20])(
    "shows up to six real reviews when there are %i reviews",
    (count) => {
      state.queries.reviews = loaded({ items: mockReviews.slice(0, count) })
      const html = renderToStaticMarkup(<HomePage />)
      expect(html.match(/<article>/g)).toHaveLength(Math.min(count, 6))
      expect(html).toContain("发现课程")
      expect(html).not.toContain("分享你上过的课程，帮助后来选课的同学。")
    }
  )

  it("distinguishes a failed request from an empty catalog and retains stale content", () => {
    state.queries.discovery = { ...loaded(undefined), isError: true }
    let html = renderToStaticMarkup(<HomePage />)
    expect(html).toContain("加载失败")
    expect(html).not.toContain("暂无课程")
    state.queries.discovery = {
      ...loaded({ semester: "2026-2027-1", items: mockCourses.slice(0, 2) }),
      isError: true,
    }
    html = renderToStaticMarkup(<HomePage />)
    expect(html).toContain("刷新失败，保留上次内容")
    expect(html).toContain(mockCourses[0].name)
    expect(html).not.toContain("换一批")
  })

  it("keeps loading and genuine empty states distinct", () => {
    state.queries.discovery = { ...loaded(undefined), isPending: true }
    let html = renderToStaticMarkup(<HomePage />)
    expect(html).toContain("正在加载内容")
    expect(html).not.toContain("暂无课程")
    state.queries.discovery = loaded({ semester: "", items: [] })
    html = renderToStaticMarkup(<HomePage />)
    expect(html).toContain("暂无课程")
    expect(html).not.toContain("换一批")
  })

  it("links a personal prompt to the existing course review form", () => {
    state.queries.prompts = loaded({
      has_enrollments: true,
      items: [{ course: mockCourses[0], semester: "2025-2026-2" }],
    })
    const html = renderToStaticMarkup(<HomePage />)
    expect(html).toContain(`href="/course/${mockCourses[0].id}/review/new"`)
    expect(html).toContain("2025-2026-2")
    expect(html).not.toContain("导入课程")
    state.queries.prompts = loaded({ has_enrollments: true, items: [] })
    expect(renderToStaticMarkup(<HomePage />)).toContain("已全部点评")
  })
})
