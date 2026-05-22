import { keepPreviousData, useQuery } from "@tanstack/react-query"
import {
  getTeacherFilters,
  listTeachers,
  listTeacherCourses,
  type TeacherListFilter,
} from "@/api/teacher"

export function useTeacherFilters() {
  return useQuery({
    queryKey: ["teacher-filters"],
    queryFn: getTeacherFilters,
  })
}

export function useTeachers(filter: TeacherListFilter = {}) {
  return useQuery({
    queryKey: ["teachers", filter],
    queryFn: () => listTeachers(filter),
    placeholderData: keepPreviousData,
  })
}

export function useTeacherCourses(
  teacherID: number,
  filter: Record<string, unknown> = {}
) {
  return useQuery({
    queryKey: ["teacher-courses", teacherID, filter],
    queryFn: () => listTeacherCourses(teacherID, filter),
    enabled: !!teacherID,
    placeholderData: keepPreviousData,
  })
}
