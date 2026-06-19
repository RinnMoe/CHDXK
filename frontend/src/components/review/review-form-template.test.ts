import { describe, expect, it } from "vitest"
import {
  DEFAULT_REVIEW_TEMPLATE,
  hasActualReviewContent,
} from "./review-form-template"

describe("hasActualReviewContent", () => {
  it("rejects empty content and unchanged template content", () => {
    expect(hasActualReviewContent("")).toBe(false)
    expect(hasActualReviewContent("\n\t ")).toBe(false)
    expect(hasActualReviewContent(DEFAULT_REVIEW_TEMPLATE)).toBe(false)
  })

  it("rejects content that only keeps template labels", () => {
    expect(
      hasActualReviewContent("课程内容：\n\n考核标准：\n\n授课质量：")
    ).toBe(false)
  })

  it("accepts prose after a template label or without a template", () => {
    expect(hasActualReviewContent("课程内容：老师讲得很清楚")).toBe(true)
    expect(hasActualReviewContent("老师讲得很清楚")).toBe(true)
  })

  it("rejects punctuation-only edits", () => {
    expect(hasActualReviewContent("课程内容：***\n\n考核标准：---")).toBe(false)
  })
})
