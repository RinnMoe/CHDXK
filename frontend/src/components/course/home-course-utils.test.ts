import { describe, expect, it } from "vitest"
import { mockCourses } from "@/mocks/fixtures/courses"
import { discoveryBatch, homeCategories } from "./home-course-utils"

describe("discovery batches", () => {
  it("cycles full and partial batches without losing courses", () => {
    const courses = mockCourses.slice(0, 14)
    expect(discoveryBatch(courses, 0)).toEqual(courses.slice(0, 6))
    expect(discoveryBatch(courses, 1)).toEqual(courses.slice(6, 12))
    expect(discoveryBatch(courses, 2)).toEqual(courses.slice(12))
    expect(discoveryBatch(courses, 3)).toEqual(courses.slice(0, 6))
    expect(discoveryBatch(courses.slice(0, 2), 100)).toHaveLength(2)
    expect(discoveryBatch([], 1)).toEqual([])
    expect(discoveryBatch(courses.slice(0, 2), 3)).toHaveLength(2)
  })
})

it("selects six populated categories by count without mutating filters", () => {
  const categories = Array.from({ length: 8 }, (_, index) => ({
    name: `category-${index}`,
    count: index,
  }))
  categories.push({ name: " ", count: 100 })
  const before = [...categories]
  expect(homeCategories(categories).map((item) => item.count)).toEqual([
    7, 6, 5, 4, 3, 2,
  ])
  expect(categories).toEqual(before)
  expect(homeCategories()).toEqual([])
})
