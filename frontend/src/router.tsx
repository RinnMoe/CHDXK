import { createBrowserRouter, Navigate } from "react-router-dom"
import { Layout } from "@/components/layout/layout"
import { CoursesPage } from "@/pages/courses-page"
import { CourseDetailPage } from "@/pages/course-detail-page"
import { LatestReviewsPage } from "@/pages/latest-reviews-page"
import { FollowedReviewsPage } from "@/pages/followed-reviews-page"
import { UserReviewsPage } from "@/pages/user-reviews-page"
import { ReviewDetailPage } from "@/pages/review-detail-page"
import { NewReviewPage } from "@/pages/new-review-page"
import { EditReviewPage } from "@/pages/edit-review-page"
import { TeachersPage } from "@/pages/teachers-page"
import { TeacherDetailPage } from "@/pages/teacher-detail-page"

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { path: "/", element: <Navigate to="/courses" replace /> },
      { path: "/courses", element: <CoursesPage /> },
      { path: "/courses/:courseID", element: <CourseDetailPage /> },
      { path: "/courses/:courseID/review/new", element: <NewReviewPage /> },
      { path: "/reviews/latest", element: <LatestReviewsPage /> },
      { path: "/reviews/followed", element: <FollowedReviewsPage /> },
      { path: "/reviews/mine", element: <UserReviewsPage /> },
      { path: "/reviews/:reviewID", element: <ReviewDetailPage /> },
      { path: "/reviews/:reviewID/edit", element: <EditReviewPage /> },
      { path: "/teachers", element: <TeachersPage /> },
      { path: "/teachers/:teacherID", element: <TeacherDetailPage /> },
    ],
  },
])
