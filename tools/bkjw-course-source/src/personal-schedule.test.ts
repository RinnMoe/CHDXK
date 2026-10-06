import { expect, it } from "vitest"
import { personalCourses } from "./personal-schedule.js"

it("deduplicates teaching sections while ignoring teacher order and opaque course ids", () => {
  expect(personalCourses([
    {courseId: "53765(26ZY1407.07)", lessonNo: "13TH1014.11", courseName: "程序设计", teacherName: "甲,乙"},
    {courseId: "53765(26ZY1407.07)", lessonNo: "13TH1014.11", courseName: "程序设计", teacherName: "乙,甲"},
    {courseId: "53800", lessonNo: "13TH1014.12", courseName: "程序设计"},
  ])).toEqual([
    {source_lesson_id: "53765", lesson_no: "13TH1014.11", name: "程序设计"},
    {source_lesson_id: "53800", lesson_no: "13TH1014.12", name: "程序设计"},
  ])
})
it("accepts an empty timetable, rejects incomplete data instead of importing partial records", () => {
  expect(personalCourses([])).toEqual([])
  for (const value of [null, {}, [{courseName: "课程"}], [{courseId: "123"}], [null]]) {
    expect(() => personalCourses(value)).toThrow("PAGE_STRUCTURE")
  }
})
