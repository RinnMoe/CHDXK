import type { SiteDailyStatDTO } from "@/api/site-stats"

function daysAgo(n: number): string {
  const d = new Date()
  d.setDate(d.getDate() - n)
  return d.toISOString().split("T")[0]
}

function nowISO(): string {
  return new Date().toISOString()
}

export function makeMockDailyStats(days: number = 30): SiteDailyStatDTO[] {
  return Array.from({ length: days }, (_, i) => ({
    stat_date: daysAgo(i),
    active_user_count: Math.floor(80 + Math.random() * 40),
    new_user_count: Math.floor(3 + Math.random() * 5),
    new_review_count: Math.floor(2 + Math.random() * 8),
    new_point_amount: Math.floor(20 + Math.random() * 80),
    review_author_count: Math.floor(5 + Math.random() * 10),
    new_like_count: Math.floor(10 + Math.random() * 20),
    new_dislike_count: Math.floor(2 + Math.random() * 5),
    total_user_count: 500 + Math.floor(Math.random() * 50),
    total_review_count: 300 + Math.floor(Math.random() * 30),
    reviewed_course_total: 120 + Math.floor(Math.random() * 10),
    generated_at: nowISO(),
    updated_at: nowISO(),
  }))
}

export function getMockYesterdayStats(): SiteDailyStatDTO {
  return makeMockDailyStats(1)[0]
}

export const mockDailyStats = makeMockDailyStats(30)
