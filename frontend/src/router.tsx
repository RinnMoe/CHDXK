import { lazy, type ComponentType } from "react"
import { createBrowserRouter } from "react-router-dom"
import { RequireAuth } from "@/components/auth/require-auth"
import { Layout } from "@/components/layout/layout"
import { PublicLayout } from "@/components/layout/public-layout"
import { RouterRoot } from "@/components/layout/route-scroll-restoration"

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
const CourseEnrollmentSyncCallbackPage = lazyNamedPage(
  () => import("@/pages/course-enrollment-sync-callback-page"),
  "CourseEnrollmentSyncCallbackPage"
)
const ApiKeysPage = lazyNamedPage(
  () => import("@/pages/api-keys-page"),
  "ApiKeysPage"
)
const UserSettingsPage = lazyNamedPage(
  () => import("@/pages/user-settings-page"),
  "UserSettingsPage"
)
const SiteStatsPage = lazyNamedPage(
  () => import("@/pages/admin/site-stats-page"),
  "SiteStatsPage"
)
const UserAdminPage = lazyNamedPage(
  () => import("@/pages/admin/user-admin-page"),
  "UserAdminPage"
)
const AboutPage = lazyNamedPage(() => import("@/pages/about-page"), "AboutPage")
const FaqPage = lazyNamedPage(() => import("@/pages/faq-page"), "FaqPage")
const NotFoundPage = lazyNamedPage(
  () => import("@/pages/not-found-page"),
  "NotFoundPage"
)

export const router = createBrowserRouter([
  {
    element: <RouterRoot />,
    children: [
      {
        element: <PublicLayout />,
        children: [
          { path: "/login", element: <LoginPage /> },
          { path: "/register", element: <RegisterPage /> },
          { path: "/password-reset", element: <PasswordResetPage /> },
          {
            path: "/course/mine/sync-callback",
            element: <CourseEnrollmentSyncCallbackPage />,
          },
        ],
      },
      {
        element: <RequireAuth />,
        children: [
          {
            element: <Layout />,
            children: [
              { path: "/", element: <HomePage /> },
              { path: "/course", element: <CoursesPage /> },
              { path: "/course/hot", element: <HotCoursesPage /> },
              { path: "/course/:courseID", element: <CourseDetailPage /> },
              {
                path: "/course/:courseID/review/new",
                element: <NewReviewPage />,
              },
              { path: "/review", element: <ReviewsPage /> },
              { path: "/review/followed", element: <FollowedReviewsPage /> },
              { path: "/review/mine", element: <UserReviewsPage /> },
              { path: "/course/mine", element: <UserCoursesPage /> },
              { path: "/review/:reviewID", element: <ReviewDetailPage /> },
              { path: "/review/:reviewID/edit", element: <EditReviewPage /> },
              { path: "/teacher", element: <TeachersPage /> },
              { path: "/teacher/:teacherID", element: <TeacherDetailPage /> },
              { path: "/point", element: <UserPointsPage /> },
              { path: "/api-key", element: <ApiKeysPage /> },
              { path: "/settings", element: <UserSettingsPage /> },
              { path: "/admin/user", element: <UserAdminPage /> },
              { path: "/admin/site-stat", element: <SiteStatsPage /> },
              { path: "/about", element: <AboutPage /> },
              { path: "/faq", element: <FaqPage /> },
              { path: "*", element: <NotFoundPage /> },
            ],
          },
        ],
      },
    ],
  },
])
