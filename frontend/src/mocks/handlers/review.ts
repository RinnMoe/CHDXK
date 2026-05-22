import { http, HttpResponse } from "msw"
import { mockReviews, findReview } from "../fixtures/reviews"
import type { ReviewDTO } from "@/api/review"

function paginate<T>(items: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    total: items.length,
    page,
    page_size: pageSize,
  }
}

function applyReviewFilter(url: URL, list: ReviewDTO[]): ReviewDTO[] {
  const semester = url.searchParams.get("semester") ?? ""
  const rating = Number(url.searchParams.get("rating") ?? "0")
  const orderBy = url.searchParams.get("order_by") ?? "created_at"
  const ascend = url.searchParams.get("ascend") === "1"

  let filtered = [...list]
  if (semester) filtered = filtered.filter((r) => r.semester === semester)
  if (rating > 0) filtered = filtered.filter((r) => r.rating === rating)

  filtered.sort((a, b) => {
    if (orderBy === "rating") {
      return ascend ? a.rating - b.rating : b.rating - a.rating
    }
    if (orderBy === "vote") {
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
  http.get("/api/review/latest", ({ request }) => {
    const url = new URL(request.url)
    const list = applyReviewFilter(url, mockReviews)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/review/followed", ({ request }) => {
    const url = new URL(request.url)
    const list = applyReviewFilter(url, mockReviews.slice(0, 12))
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(list, page, pageSize))
  }),

  http.get("/api/review/:reviewID", ({ params }) => {
    const id = Number(params.reviewID)
    const review = findReview(id)
    if (!review) {
      return HttpResponse.json({ error: "review not found" }, { status: 404 })
    }
    return HttpResponse.json(review)
  }),

  http.post("/api/review/", async ({ request }) => {
    await request.json()
    return HttpResponse.json({ message: "ok" }, { status: 201 })
  }),

  http.put("/api/review/:reviewID", async ({ request }) => {
    await request.json()
    return HttpResponse.json({ message: "ok" })
  }),

  http.delete("/api/review/:reviewID", () => {
    return HttpResponse.json({ message: "ok" })
  }),

  http.post("/api/review/:reviewID/vote", async ({ request, params }) => {
    const id = Number(params.reviewID)
    const review = findReview(id)
    const body = (await request.json()) as { vote_type: number }
    if (review) {
      review.vote.my_vote = body.vote_type
    }
    return HttpResponse.json({ message: "ok" })
  }),

  http.get("/api/user/:userID/reviews", ({ params, request }) => {
    const userID = Number(params.userID)
    const url = new URL(request.url)
    // mock: first 8 reviews belong to userID 1
    const list = userID === 1 ? mockReviews.slice(0, 8) : []
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")
    return HttpResponse.json(paginate(applyReviewFilter(url, list), page, pageSize))
  }),
]
