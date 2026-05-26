import { http, HttpResponse } from "msw"
import { mockReviews, findReview } from "../fixtures/reviews"
import { findUserByID, mockSession } from "../fixtures/auth"
import { randomDelay } from "../utils"
import type { ReviewDTO, ReviewRevisionDTO } from "@/api/review"

const DEFAULT_PAGE = 1
const DEFAULT_PAGE_SIZE = 20
const MAX_PAGE_SIZE = 100

function normalizePage(value: number) {
  return value > 0 ? value : DEFAULT_PAGE
}

function normalizePageSize(value: number) {
  if (value <= 0) return DEFAULT_PAGE_SIZE
  return Math.min(value, MAX_PAGE_SIZE)
}

function paginate<T>(items: T[], page: number, pageSize: number) {
  page = normalizePage(page)
  pageSize = normalizePageSize(pageSize)
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    total: items.length,
    page,
    page_size: pageSize,
  }
}

function withPrivateFields(review: ReviewDTO): ReviewDTO {
  const user = mockSession.userID ? findUserByID(mockSession.userID) : undefined
  if (!user) return { ...review, user_id: undefined }
  if (review.user_id === user.id || user.role === "admin") return { ...review }
  return { ...review, user_id: undefined }
}

function buildRevisions(review: ReviewDTO): ReviewRevisionDTO[] {
  const createdAt = new Date(review.created_at).getTime()
  const updatedAt = new Date(review.updated_at).getTime()
  if (updatedAt <= createdAt + 1000) {
    return []
  }
  const revisionCount = 2 + (review.id % 3)
  const step = Math.max(60 * 60 * 1000, Math.floor((updatedAt - createdAt) / 4))

  return Array.from({ length: revisionCount }, (_, i) => {
    const revisionTime = Math.max(createdAt + 1000, updatedAt - i * step)
    return {
      id: review.id * 100 + i,
      review_id: review.id,
      course_id: review.course_id,
      created_by: review.user_id ?? 0,
      semester: review.semester ?? "",
      score: review.score,
      rating: Math.max(1, review.rating - i),
      content: review.content,
      created_at: new Date(revisionTime).toISOString(),
    }
  })
}

function isCurrentUserAdmin() {
  const user = mockSession.userID ? findUserByID(mockSession.userID) : undefined
  return user?.role === "admin"
}

function applyReviewFilter(url: URL, list: ReviewDTO[]): ReviewDTO[] {
  const q = (url.searchParams.get("q") ?? "").trim().toLowerCase()
  const semester = url.searchParams.get("semester") ?? ""
  const rating = Number(url.searchParams.get("rating") ?? "0")
  const orderBy = url.searchParams.get("order_by") ?? "created_at"
  const ascend = url.searchParams.get("ascend") === "1"

  let filtered = [...list]
  if (q) filtered = filtered.filter((r) => r.content.toLowerCase().includes(q))
  if (semester) filtered = filtered.filter((r) => r.semester === semester)
  if (rating > 0) filtered = filtered.filter((r) => r.rating === rating)

  filtered.sort((a, b) => {
    if (orderBy === "rating") {
      return ascend ? a.rating - b.rating : b.rating - a.rating
    }
    if (orderBy === "like_count") {
      return ascend
        ? a.vote.like_count - b.vote.like_count
        : b.vote.like_count - a.vote.like_count
    }
    const aT = new Date(a.created_at).getTime()
    const bT = new Date(b.created_at).getTime()
    return ascend ? aT - bT : bT - aT
  })
  return filtered
}

export const reviewHandlers = [
  http.get("/api/review", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const list = applyReviewFilter(url, mockReviews).map(withPrivateFields)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/review/followed", async ({ request }) => {
    await randomDelay()
    const url = new URL(request.url)
    const list = applyReviewFilter(url, mockReviews.slice(0, 12)).map(
      withPrivateFields
    )
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/review/:reviewID/revision", async ({ params }) => {
    await randomDelay()
    if (!isCurrentUserAdmin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    const id = Number(params.reviewID)
    const review = findReview(id)
    if (!review) {
      return HttpResponse.json({ error: "review not found" }, { status: 404 })
    }
    return HttpResponse.json(buildRevisions(review))
  }),

  http.get("/api/review/:reviewID", async ({ params }) => {
    await randomDelay()
    const id = Number(params.reviewID)
    const review = findReview(id)
    if (!review) {
      return HttpResponse.json({ error: "review not found" }, { status: 404 })
    }
    return HttpResponse.json(withPrivateFields(review))
  }),

  http.post("/api/review/", async ({ request }) => {
    await randomDelay()
    await request.json()
    return HttpResponse.json({ message: "ok" }, { status: 201 })
  }),

  http.put("/api/review/:reviewID", async ({ request }) => {
    await randomDelay()
    await request.json()
    return HttpResponse.json({ message: "ok" })
  }),

  http.put(
    "/api/review/:reviewID/moderator-remark",
    async ({ request, params }) => {
      await randomDelay()
      if (!isCurrentUserAdmin()) {
        return HttpResponse.json({ error: "forbidden" }, { status: 403 })
      }
      const id = Number(params.reviewID)
      const review = findReview(id)
      if (!review) {
        return HttpResponse.json({ error: "review not found" }, { status: 404 })
      }
      const body = (await request.json()) as { moderator_remark?: string }
      review.moderator_remark = body.moderator_remark ?? ""
      return HttpResponse.json({ message: "ok" })
    }
  ),

  http.delete("/api/review/:reviewID", async ({ params }) => {
    await randomDelay()
    if (!isCurrentUserAdmin()) {
      return HttpResponse.json({ error: "forbidden" }, { status: 403 })
    }
    const id = Number(params.reviewID)
    if (!findReview(id)) {
      return HttpResponse.json({ error: "review not found" }, { status: 404 })
    }
    return HttpResponse.json({ message: "ok" })
  }),

  http.post("/api/review/:reviewID/vote", async ({ request, params }) => {
    await randomDelay()
    const id = Number(params.reviewID)
    const review = findReview(id)
    const body = (await request.json()) as { vote_type: number }
    if (review) {
      review.vote.my_vote = body.vote_type
    }
    return HttpResponse.json({ message: "ok" })
  }),

  http.get("/api/user/:userID/review", async ({ params, request }) => {
    await randomDelay()
    const userID = Number(params.userID)
    const url = new URL(request.url)
    // mock: first 8 reviews belong to userID 1
    const list =
      userID === 1 ? mockReviews.slice(0, 8).map(withPrivateFields) : []
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(
      paginate(applyReviewFilter(url, list), page, pageSize)
    )
  }),
]
