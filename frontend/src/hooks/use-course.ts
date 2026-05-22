import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  listCourses,
  getCourseFilters,
  getCourseDetail,
  listCourseReviews,
  getCourseReviewFilters,
  setNotificationLevel,
  listFollowedCourses,
  listIgnoredCourses,
  type CourseListFilter,
} from "@/api/course"
import type { ReviewListFilter } from "@/api/review"

export function useCourseFilters() {
  return useQuery({
    queryKey: ["course-filters"],
    queryFn: getCourseFilters,
  })
}

export function useCourses(filter: CourseListFilter = {}) {
  return useQuery({
    queryKey: ["courses", filter],
    queryFn: () => listCourses(filter),
  })
}

export function useCourseDetail(courseID: number) {
  return useQuery({
    queryKey: ["course", courseID],
    queryFn: () => getCourseDetail(courseID),
    enabled: !!courseID,
  })
}

export function useCourseReviews(
  courseID: number,
  filter: ReviewListFilter = {}
) {
  return useQuery({
    queryKey: ["course-reviews", courseID, filter],
    queryFn: () => listCourseReviews(courseID, filter),
    enabled: !!courseID,
  })
}

export function useCourseReviewFilters(courseID: number) {
  return useQuery({
    queryKey: ["course-review-filters", courseID],
    queryFn: () => getCourseReviewFilters(courseID),
    enabled: !!courseID,
  })
}

export function useSetNotificationLevel() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ courseID, level }: { courseID: number; level: number }) =>
      setNotificationLevel(courseID, level),
    onSuccess: (_, { courseID }) => {
      queryClient.invalidateQueries({ queryKey: ["course", courseID] })
    },
  })
}

export function useFollowedCourses(filter: CourseListFilter = {}) {
  return useQuery({
    queryKey: ["followed-courses", filter],
    queryFn: () => listFollowedCourses(filter),
  })
}

export function useIgnoredCourses(filter: CourseListFilter = {}) {
  return useQuery({
    queryKey: ["ignored-courses", filter],
    queryFn: () => listIgnoredCourses(filter),
  })
}
