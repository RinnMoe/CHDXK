import { keepPreviousData, useQuery } from "@tanstack/react-query"
import {
  getTeacher,
  getTeacherFilters,
  listTeachers,
  listTeacherCourses,
  listHotTeachers,
  type TeacherListFilter,
} from "@/api/teacher"

export function useHotTeachers(period: "week" | "month", limit = 50) {
  return useQuery({
    queryKey: ["hot-teachers", period, limit],
    queryFn: () => listHotTeachers(period, limit),
  })
}

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

export function useTeacher(teacherID: number) {
  return useQuery({
    queryKey: ["teacher", teacherID],
    queryFn: () => getTeacher(teacherID),
    enabled: !!teacherID,
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
