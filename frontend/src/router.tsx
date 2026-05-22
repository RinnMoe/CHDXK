import { createBrowserRouter } from "react-router-dom"
import { Layout } from "@/components/layout/layout"
import { HomePage } from "@/pages/home-page"
import { CoursesPage } from "@/pages/courses-page"
import { HotCoursesPage } from "@/pages/hot-courses-page"
import { CourseDetailPage } from "@/pages/course-detail-page"
import { LatestReviewsPage } from "@/pages/latest-reviews-page"
import { FollowedReviewsPage } from "@/pages/followed-reviews-page"
import { UserReviewsPage } from "@/pages/user-reviews-page"
import { ReviewDetailPage } from "@/pages/review-detail-page"
import { NewReviewPage } from "@/pages/new-review-page"
import { EditReviewPage } from "@/pages/edit-review-page"
import { TeachersPage } from "@/pages/teachers-page"
import { TeacherDetailPage } from "@/pages/teacher-detail-page"
import { LoginPage } from "@/pages/login-page"
import { RegisterPage } from "@/pages/register-page"
import { PasswordResetPage } from "@/pages/password-reset-page"
import { UserPointsPage } from "@/pages/user-points-page"
import { UserCoursesPage } from "@/pages/user-courses-page"
import { SiteStatsPage } from "@/pages/admin/site-stats-page"

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/courses", element: <CoursesPage /> },
      { path: "/courses/hot", element: <HotCoursesPage /> },
      { path: "/courses/:courseID", element: <CourseDetailPage /> },
      { path: "/courses/:courseID/review/new", element: <NewReviewPage /> },
      { path: "/reviews/latest", element: <LatestReviewsPage /> },
      { path: "/reviews/followed", element: <FollowedReviewsPage /> },
      { path: "/reviews/mine", element: <UserReviewsPage /> },
      { path: "/courses/mine", element: <UserCoursesPage /> },
      { path: "/reviews/:reviewID", element: <ReviewDetailPage /> },
      { path: "/reviews/:reviewID/edit", element: <EditReviewPage /> },
      { path: "/teachers", element: <TeachersPage /> },
      { path: "/teachers/:teacherID", element: <TeacherDetailPage /> },
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/password-reset", element: <PasswordResetPage /> },
      { path: "/points", element: <UserPointsPage /> },
      { path: "/admin/site-stats", element: <SiteStatsPage /> },
    ],
  },
])
