import type { TeacherDTO } from "@/api/teacher"
import { mockCourses } from "./courses"

const TITLES = ["教授", "副教授", "讲师", "助理教授", "副研究员"]

function randInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function pick<T>(arr: T[]): T {
  return arr[randInt(0, arr.length - 1)]
}

// Re-use teachers from mock courses to keep data consistent
const teacherMap = new Map<number, TeacherDTO>()

export function getMockTeachers(): TeacherDTO[] {
  if (teacherMap.size > 0) return Array.from(teacherMap.values())

  const seen = new Set<number>()
  for (const course of mockCourses) {
    if (seen.has(course.main_teacher.id)) continue
    seen.add(course.main_teacher.id)
    teacherMap.set(course.main_teacher.id, {
      id: course.main_teacher.id,
      code: `T${String(course.main_teacher.id).padStart(5, "0")}`,
      name: course.main_teacher.name,
      department: course.main_teacher.department,
      title: pick(TITLES),
    })
  }
  return Array.from(teacherMap.values())
}

export function getMockTeacher(id: number): TeacherDTO | undefined {
  return getMockTeachers().find((t) => t.id === id)
}

export function makeTeacherFilters() {
  const teachers = getMockTeachers()
  const deptMap = new Map<string, number>()
  const titleMap = new Map<string, number>()
  for (const t of teachers) {
    deptMap.set(t.department, (deptMap.get(t.department) ?? 0) + 1)
    if (t.title) titleMap.set(t.title, (titleMap.get(t.title) ?? 0) + 1)
  }
  return {
    departments: Array.from(deptMap.entries()).map(([name, count]) => ({
      name,
      count,
    })),
    titles: Array.from(titleMap.entries()).map(([name, count]) => ({
      name,
      count,
    })),
  }
}
