import type { QueryClient } from "@tanstack/react-query"
import {
  createRootRouteWithContext,
  createRoute,
  createRouter,
  lazyRouteComponent,
  Outlet,
  redirect,
} from "@tanstack/react-router"

import { Layout } from "@/components/layout/layout"
import { PublicLayout } from "@/components/layout/public-layout"
import { RouteErrorPage } from "@/components/layout/route-error-page"
import { RouteFallback } from "@/components/layout/route-fallback"
import { getSafeRedirectPath } from "@/lib/auth-redirect"
import {
  validateCourseDetailSearch,
  validateCourseEnrollmentSyncCallbackSearch,
  validateCourseListSearch,
  validateLoginSearch,
  validatePageSearch,
  validateReviewsSearch,
  validateSiteStatsSearch,
  validateTeacherDetailSearch,
  validateTeachersSearch,
  validateUserAdminSearch,
  validateUserCoursesSearch,
} from "./router-search"
import { authMeQueryOptions } from "@/hooks/use-auth"

type RouterContext = {
  queryClient: QueryClient
}

export type {
  CourseDetailSearch,
  CourseEnrollmentSyncCallbackSearch,
  CourseListSearch,
  LoginSearch,
  PageSearch,
  ReviewsSearch,
  SiteStatsSearch,
  TeacherDetailSearch,
  TeachersSearch,
  UserAdminSearch,
  UserCoursesSearch,
} from "./router-search"

function buildLoginRedirectPath(path: string) {
  return `/login?${new URLSearchParams({ redirect: path }).toString()}`
}

function currentPath(location: {
  pathname: string
  searchStr: string
  hash: string
}) {
  return `${location.pathname}${location.searchStr}${location.hash}`
}

async function requireAuth({
  context,
  location,
}: {
  context: RouterContext
  location: { pathname: string; searchStr: string; hash: string }
}) {
  const user = await context.queryClient.ensureQueryData(authMeQueryOptions())
  if (!user) {
    throw redirect({
      to: "/login",
      search: { redirect: currentPath(location) },
      replace: true,
    })
  }
  return { user }
}

async function redirectAuthedUser({
  context,
  search,
}: {
  context: RouterContext
  search?: { redirect?: string }
}) {
  const user = await context.queryClient.ensureQueryData(authMeQueryOptions())
  if (user) {
    throw redirect({
      to: getSafeRedirectPath(search?.redirect),
      replace: true,
    })
  }
}

const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: Outlet,
  errorComponent: RouteErrorPage,
})

const publicRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "public",
  component: PublicLayout,
})

const appRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "app",
  beforeLoad: requireAuth,
  component: Layout,
})

export const homeRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/",
  component: lazyRouteComponent(() => import("@/pages/home-page"), "HomePage"),
})

export const loginRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/login",
  validateSearch: validateLoginSearch,
  beforeLoad: redirectAuthedUser,
  component: lazyRouteComponent(() => import("@/pages/login-page"), "LoginPage"),
})

export const registerRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/register",
  beforeLoad: redirectAuthedUser,
  component: lazyRouteComponent(() => import("@/pages/register-page"), "RegisterPage"),
})

export const passwordResetRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/password-reset",
  beforeLoad: redirectAuthedUser,
  component: lazyRouteComponent(
    () => import("@/pages/password-reset-page"),
    "PasswordResetPage"
  ),
})

export const courseEnrollmentSyncCallbackRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/course/mine/sync-callback",
  validateSearch: validateCourseEnrollmentSyncCallbackSearch,
  component: lazyRouteComponent(
    () => import("@/pages/course-enrollment-sync-callback-page"),
    "CourseEnrollmentSyncCallbackPage"
  ),
})

export const coursesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course",
  validateSearch: validateCourseListSearch,
  component: lazyRouteComponent(() => import("@/pages/courses-page"), "CoursesPage"),
})

export const hotCoursesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course/hot",
  component: lazyRouteComponent(() => import("@/pages/hot-courses-page"), "HotCoursesPage"),
})

export const courseDetailRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course/$courseID",
  validateSearch: validateCourseDetailSearch,
  component: lazyRouteComponent(
    () => import("@/pages/course-detail-page"),
    "CourseDetailPage"
  ),
})

export const newReviewRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course/$courseID/review/new",
  component: lazyRouteComponent(
    () => import("@/pages/new-review-page"),
    "NewReviewPage"
  ),
})

export const reviewsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review",
  validateSearch: validateReviewsSearch,
  component: lazyRouteComponent(
    () => import("@/pages/reviews-page"),
    "ReviewsPage"
  ),
})

export const followedReviewsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review/followed",
  validateSearch: validatePageSearch,
  component: lazyRouteComponent(
    () => import("@/pages/followed-reviews-page"),
    "FollowedReviewsPage"
  ),
})

export const userReviewsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review/mine",
  validateSearch: validatePageSearch,
  component: lazyRouteComponent(
    () => import("@/pages/user-reviews-page"),
    "UserReviewsPage"
  ),
})

export const reviewDetailRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review/$reviewID",
  component: lazyRouteComponent(
    () => import("@/pages/review-detail-page"),
    "ReviewDetailPage"
  ),
})

export const editReviewRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review/$reviewID/edit",
  component: lazyRouteComponent(
    () => import("@/pages/edit-review-page"),
    "EditReviewPage"
  ),
})

export const userCoursesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course/mine",
  validateSearch: validateUserCoursesSearch,
  component: lazyRouteComponent(
    () => import("@/pages/user-courses-page"),
    "UserCoursesPage"
  ),
})

export const teachersRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/teacher",
  validateSearch: validateTeachersSearch,
  component: lazyRouteComponent(
    () => import("@/pages/teachers-page"),
    "TeachersPage"
  ),
})

export const teacherDetailRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/teacher/$teacherID",
  validateSearch: validateTeacherDetailSearch,
  component: lazyRouteComponent(
    () => import("@/pages/teacher-detail-page"),
    "TeacherDetailPage"
  ),
})

export const userPointsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/point",
  component: lazyRouteComponent(
    () => import("@/pages/user-points-page"),
    "UserPointsPage"
  ),
})

export const apiKeysRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/api-key",
  component: lazyRouteComponent(
    () => import("@/pages/api-keys-page"),
    "ApiKeysPage"
  ),
})

export const userSettingsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/settings",
  component: lazyRouteComponent(
    () => import("@/pages/user-settings-page"),
    "UserSettingsPage"
  ),
})

export const userAdminRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/admin/user",
  validateSearch: validateUserAdminSearch,
  component: lazyRouteComponent(
    () => import("@/pages/admin/user-admin-page"),
    "UserAdminPage"
  ),
})

export const siteStatsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/admin/site-stat",
  validateSearch: validateSiteStatsSearch,
  component: lazyRouteComponent(
    () => import("@/pages/admin/site-stats-page"),
    "SiteStatsPage"
  ),
})

export const aboutRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/about",
  component: lazyRouteComponent(
    () => import("@/pages/about-page"),
    "AboutPage"
  ),
})

export const faqRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/faq",
  component: lazyRouteComponent(() => import("@/pages/faq-page"), "FaqPage"),
})

const latestRedirectRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/latest",
  beforeLoad: () => {
    throw redirect({ to: "/review", replace: true })
  },
})

const notFoundRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "$",
  component: lazyRouteComponent(
    () => import("@/pages/not-found-page"),
    "NotFoundPage"
  ),
})

const routeTree = rootRoute.addChildren([
  publicRoute.addChildren([
    loginRoute,
    registerRoute,
    passwordResetRoute,
    courseEnrollmentSyncCallbackRoute,
  ]),
  appRoute.addChildren([
    homeRoute,
    coursesRoute,
    hotCoursesRoute,
    courseDetailRoute,
    newReviewRoute,
    reviewsRoute,
    followedReviewsRoute,
    userReviewsRoute,
    userCoursesRoute,
    reviewDetailRoute,
    editReviewRoute,
    teachersRoute,
    teacherDetailRoute,
    userPointsRoute,
    apiKeysRoute,
    userSettingsRoute,
    userAdminRoute,
    siteStatsRoute,
    aboutRoute,
    faqRoute,
    latestRedirectRoute,
    notFoundRoute,
  ]),
])

export const router = createRouter({
  routeTree,
  context: {
    queryClient: undefined!,
  },
  defaultPendingComponent: RouteFallback,
  defaultPendingMs: 250,
  scrollRestoration: true,
})

export function getLoginRedirectPath() {
  return buildLoginRedirectPath(
    `${router.state.location.pathname}${router.state.location.searchStr}${router.state.location.hash}`
  )
}

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
