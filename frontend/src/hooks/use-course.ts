import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query"
import {
  listCourses,
  getCourseFilters,
  getCourseDetail,
  listCourseReviews,
  getCourseReviewFilters,
  setNotificationLevel,
  listFollowedCourses,
  listIgnoredCourses,
  type CourseDetailDTO,
  type CourseListFilter,
  type CourseNotificationLevel,
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
    placeholderData: keepPreviousData,
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
    placeholderData: keepPreviousData,
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
    mutationFn: ({
      courseID,
      level,
    }: {
      courseID: number
      level: CourseNotificationLevel
    }) => setNotificationLevel(courseID, level),
    onMutate: async ({ courseID, level }) => {
      await queryClient.cancelQueries({ queryKey: ["course", courseID] })
      const previous = queryClient.getQueryData<CourseDetailDTO>([
        "course",
        courseID,
      ])
      queryClient.setQueryData<CourseDetailDTO>(["course", courseID], (course) =>
        course ? { ...course, notification_level: level } : course
      )
      return { previous, courseID }
    },
    onError: (_err, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["course", context.courseID], context.previous)
      }
    },
    onSuccess: (_, { courseID }) => {
      queryClient.invalidateQueries({ queryKey: ["course", courseID] })
      queryClient.invalidateQueries({ queryKey: ["followed-courses"] })
      queryClient.invalidateQueries({ queryKey: ["ignored-courses"] })
    },
  })
}

export function useFollowedCourses(
  filter: CourseListFilter = {},
  enabled = true
) {
  return useQuery({
    queryKey: ["followed-courses", filter],
    queryFn: () => listFollowedCourses(filter),
    enabled,
    placeholderData: keepPreviousData,
  })
}

export function useIgnoredCourses(filter: CourseListFilter = {}, enabled = true) {
  return useQuery({
    queryKey: ["ignored-courses", filter],
    queryFn: () => listIgnoredCourses(filter),
    enabled,
    placeholderData: keepPreviousData,
  })
}
