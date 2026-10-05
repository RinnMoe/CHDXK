import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  createManagedCourse,
  deleteManagedCourse,
  listManagedCourses,
  listManagedCourseSemesters,
  updateManagedCourse,
  type AdminCourseFilter,
  type SaveManagedCourseOffering,
} from "@/api/admin-course"

export async function invalidateCourseDataQueries(
  queryClient: ReturnType<typeof useQueryClient>
) {
  const queryKeys = [
    ["admin-courses"],
    ["course"],
    ["courses"],
    ["course-filters"],
    ["course-reviews"],
    ["course-review-filters"],
    ["course-review-trend"],
    ["teachers"],
    ["teacher"],
    ["teacher-courses"],
    ["teacher-filters"],
  ]
  await Promise.all(
    queryKeys.map((queryKey) => queryClient.invalidateQueries({ queryKey }))
  )
}

export function useManagedCourseSemesters() {
  return useQuery({
    queryKey: ["admin-courses", "semesters"],
    queryFn: listManagedCourseSemesters,
  })
}

export function useManagedCourses(filter: AdminCourseFilter, enabled = true) {
  return useQuery({
    queryKey: ["admin-courses", filter],
    queryFn: () => listManagedCourses(filter),
    enabled: enabled && filter.semester !== "",
  })
}

export function useSaveManagedCourse() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      command,
    }: {
      id?: number
      command: SaveManagedCourseOffering
    }) =>
      id ? updateManagedCourse(id, command) : createManagedCourse(command),
    onSuccess: () => invalidateCourseDataQueries(queryClient),
  })
}

export function useDeleteManagedCourse() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteManagedCourse,
    onSuccess: () => invalidateCourseDataQueries(queryClient),
  })
}
