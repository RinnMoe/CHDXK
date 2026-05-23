import { lazy, type ComponentType } from "react"
import { createBrowserRouter } from "react-router-dom"
import { RequireAuth } from "@/components/auth/require-auth"
import { Layout } from "@/components/layout/layout"
import { PublicLayout } from "@/components/layout/public-layout"

function lazyNamedPage<TModule extends Record<string, ComponentType>>(
  loader: () => Promise<TModule>,
  exportName: keyof TModule
) {
  return lazy(() =>
    loader().then((module) => ({
      default: module[exportName],
    }))
  )
}

const HomePage = lazyNamedPage(() => import("@/pages/home-page"), "HomePage")
const CoursesPage = lazyNamedPage(
  () => import("@/pages/courses-page"),
  "CoursesPage"
)
const HotCoursesPage = lazyNamedPage(
  () => import("@/pages/hot-courses-page"),
  "HotCoursesPage"
)
const CourseDetailPage = lazyNamedPage(
  () => import("@/pages/course-detail-page"),
  "CourseDetailPage"
)
const ReviewsPage = lazyNamedPage(
  () => import("@/pages/reviews-page"),
  "ReviewsPage"
)
const FollowedReviewsPage = lazyNamedPage(
  () => import("@/pages/followed-reviews-page"),
  "FollowedReviewsPage"
)
const UserReviewsPage = lazyNamedPage(
  () => import("@/pages/user-reviews-page"),
  "UserReviewsPage"
)
const ReviewDetailPage = lazyNamedPage(
  () => import("@/pages/review-detail-page"),
  "ReviewDetailPage"
)
const NewReviewPage = lazyNamedPage(
  () => import("@/pages/new-review-page"),
  "NewReviewPage"
)
const EditReviewPage = lazyNamedPage(
  () => import("@/pages/edit-review-page"),
  "EditReviewPage"
)
const TeachersPage = lazyNamedPage(
  () => import("@/pages/teachers-page"),
  "TeachersPage"
)
const TeacherDetailPage = lazyNamedPage(
  () => import("@/pages/teacher-detail-page"),
  "TeacherDetailPage"
)
const LoginPage = lazyNamedPage(() => import("@/pages/login-page"), "LoginPage")
const RegisterPage = lazyNamedPage(
  () => import("@/pages/register-page"),
  "RegisterPage"
)
const PasswordResetPage = lazyNamedPage(
  () => import("@/pages/password-reset-page"),
  "PasswordResetPage"
)
const UserPointsPage = lazyNamedPage(
  () => import("@/pages/user-points-page"),
  "UserPointsPage"
)
const UserCoursesPage = lazyNamedPage(
  () => import("@/pages/user-courses-page"),
  "UserCoursesPage"
)
const ApiKeysPage = lazyNamedPage(
  () => import("@/pages/api-keys-page"),
  "ApiKeysPage"
)
const SiteStatsPage = lazyNamedPage(
  () => import("@/pages/admin/site-stats-page"),
  "SiteStatsPage"
)
const AboutPage = lazyNamedPage(() => import("@/pages/about-page"), "AboutPage")
const FaqPage = lazyNamedPage(() => import("@/pages/faq-page"), "FaqPage")
const NotFoundPage = lazyNamedPage(
  () => import("@/pages/not-found-page"),
  "NotFoundPage"
)

export const router = createBrowserRouter([
  {
    element: <PublicLayout />,
    children: [
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/password-reset", element: <PasswordResetPage /> },
    ],
  },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <Layout />,
        children: [
          { path: "/", element: <HomePage /> },
          { path: "/courses", element: <CoursesPage /> },
          { path: "/courses/hot", element: <HotCoursesPage /> },
          { path: "/courses/:courseID", element: <CourseDetailPage /> },
          { path: "/courses/:courseID/review/new", element: <NewReviewPage /> },
          { path: "/reviews", element: <ReviewsPage /> },
          { path: "/reviews/followed", element: <FollowedReviewsPage /> },
          { path: "/reviews/mine", element: <UserReviewsPage /> },
          { path: "/courses/mine", element: <UserCoursesPage /> },
          { path: "/reviews/:reviewID", element: <ReviewDetailPage /> },
          { path: "/reviews/:reviewID/edit", element: <EditReviewPage /> },
          { path: "/teachers", element: <TeachersPage /> },
          { path: "/teachers/:teacherID", element: <TeacherDetailPage /> },
          { path: "/points", element: <UserPointsPage /> },
          { path: "/api-keys", element: <ApiKeysPage /> },
          { path: "/admin/site-stats", element: <SiteStatsPage /> },
          { path: "/about", element: <AboutPage /> },
          { path: "/faq", element: <FaqPage /> },
          { path: "*", element: <NotFoundPage /> },
        ],
      },
    ],
  },
])
