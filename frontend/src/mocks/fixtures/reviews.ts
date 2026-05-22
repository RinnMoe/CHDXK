import type { ReviewDTO } from "@/api/review"
import { mockCourses } from "./courses"

const SCORES = ["A+", "A", "A-", "B+", "B", "B-", "C+", "C", "未公布"]
const SEMESTERS = ["2025-2026-1", "2024-2025-2", "2024-2025-1", "2023-2024-2"]

const SAMPLE_CONTENTS = [
  "讲得非常清楚，作业难度适中，考试也比较公平。",
  "课程内容硬核，需要花大量时间预习和复习。老师水平很高，但对学生要求也高。",
  "比较水的课，平时点名加期末小论文，不挂人。适合刷绩点。",
  "推荐！老师人很好，讲课节奏适中，PPT 也做得很认真。",
  "课程难度偏高，但内容很有意思。注意期中考试占 40%，平时一定要跟上。",
  "授课方式比较传统，主要是 PPT 念读。建议自学为主。",
  "老师讲得很有激情，但内容太多，每节课信息量爆炸。考试前需要好好整理笔记。",
  "总体不错，作业有些多但都很有针对性。期末考试题目不算难，但需要灵活应用。",
]

function randInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function pick<T>(arr: T[]): T {
  return arr[randInt(0, arr.length - 1)]
}

let reviewIDSeed = 1

export function generateReviewsForCourse(courseID: number, n = 8): ReviewDTO[] {
  const course = mockCourses.find((c) => c.id === courseID)
  return Array.from({ length: n }, () => {
    const created = new Date(Date.now() - randInt(1, 365) * 86400000)
    // ~30% chance the review was edited later
    const updated =
      randInt(0, 9) < 3
        ? new Date(
            created.getTime() + randInt(1, 30) * 86400000 + randInt(0, 86400) * 1000
          )
        : created
    return {
      id: reviewIDSeed++,
      course_id: courseID,
      course,
      semester: pick(SEMESTERS),
      score: pick(SCORES),
      rating: randInt(1, 5),
      content: pick(SAMPLE_CONTENTS),
      vote: {
        like_count: randInt(0, 30),
        dislike_count: randInt(0, 5),
      },
      created_at: created.toISOString(),
      updated_at: updated.toISOString(),
    } as ReviewDTO
  })
}

export const mockReviews: ReviewDTO[] = mockCourses.flatMap((c) =>
  generateReviewsForCourse(c.id, randInt(0, 6))
)

export function findReview(id: number): ReviewDTO | undefined {
  return mockReviews.find((r) => r.id === id)
}
