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
import { RouteFallback } from "@/components/layout/route-fallback"
import { getSafeRedirectPath } from "@/lib/auth-redirect"
import {
  enumParam,
  numberParam,
  positiveIntParam,
  rawStringParam,
  stringArrayParam,
  stringParam,
  type SearchRecord,
} from "@/lib/router/search"
import { authMeQueryOptions } from "@/hooks/use-auth"

type RouterContext = {
  queryClient: QueryClient
}

export type LoginSearch = { redirect?: string }
export type CourseListSearch = {
  q?: string
  department?: string
  language?: string
  categories?: string[]
  target_years?: string[]
  credit?: number
  order_by?: "rating_score" | "rating_count"
  page?: number
}
export type CourseDetailSearch = {
  page?: number
  semester?: string
  rating?: number
  order_by?: "created_at" | "updated_at" | "like_count"
}
export type PageSearch = { page?: number }
export type ReviewsSearch = PageSearch & { q?: string }
export type UserCoursesSearch = PageSearch & {
  type?: "enrolled" | "followed" | "ignored"
  semester?: string
}
export type TeachersSearch = PageSearch & {
  department?: string
  title?: string
  q?: string
}
export type TeacherDetailSearch = PageSearch & {
  order_by?: "rating_score" | "rating_count"
}
export type UserAdminSearch = {
  tab?: "user" | "admin" | "system-api-key" | "audit-log"
  email?: string
  page?: number
  audit_page?: number
  audit_start_time?: string
  audit_end_time?: string
  audit_action?: string
  audit_actor_user_id?: number
}
export type SiteStatsSearch = { start_date?: string; end_date?: string }
export type CourseEnrollmentSyncCallbackSearch = {
  status?: "ok" | "error"
  semester?: string
  message?: string
  matched?: number
  total?: number
}

function buildLoginRedirectPath(path: string) {
  return `/login?${new URLSearchParams({ redirect: path }).toString()}`
}

function currentPath(location: { pathname: string; searchStr: string; hash: string }) {
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
  validateSearch: (search: SearchRecord): LoginSearch => ({
    redirect: rawStringParam(search.redirect)
      ? getSafeRedirectPath(rawStringParam(search.redirect))
      : undefined,
  }),
  beforeLoad: redirectAuthedUser,
  component: lazyRouteComponent(() => import("@/pages/login-page"), "LoginPage"),
})

export const registerRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/register",
  beforeLoad: redirectAuthedUser,
  component: lazyRouteComponent(
    () => import("@/pages/register-page"),
    "RegisterPage"
  ),
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
  validateSearch: (search: SearchRecord): CourseEnrollmentSyncCallbackSearch => ({
    status: enumParam(search.status, ["ok", "error"] as const) ?? "error",
    semester: rawStringParam(search.semester) ?? "",
    message: rawStringParam(search.message),
    matched: numberParam(search.matched) ?? 0,
    total: numberParam(search.total) ?? 0,
  }),
  component: lazyRouteComponent(
    () => import("@/pages/course-enrollment-sync-callback-page"),
    "CourseEnrollmentSyncCallbackPage"
  ),
})

export const coursesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course",
  validateSearch: (search: SearchRecord): CourseListSearch => ({
    q: stringParam(search.q),
    department: rawStringParam(search.department),
    language: rawStringParam(search.language),
    categories: stringArrayParam(search.categories),
    target_years: stringArrayParam(search.target_years),
    credit: numberParam(search.credit),
    order_by: enumParam(search.order_by, [
      "rating_score",
      "rating_count",
    ] as const),
    page: numberParam(search.page),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/courses-page"),
    "CoursesPage"
  ),
})

export const hotCoursesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course/hot",
  component: lazyRouteComponent(
    () => import("@/pages/hot-courses-page"),
    "HotCoursesPage"
  ),
})

export const courseDetailRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course/$courseID",
  validateSearch: (search: SearchRecord): CourseDetailSearch => ({
    page: numberParam(search.page),
    semester: rawStringParam(search.semester),
    rating: numberParam(search.rating),
    order_by: enumParam(search.order_by, [
      "created_at",
      "updated_at",
      "like_count",
    ] as const),
  }),
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
  validateSearch: (search: SearchRecord): ReviewsSearch => ({
    q: stringParam(search.q),
    page: numberParam(search.page),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/reviews-page"),
    "ReviewsPage"
  ),
})

export const followedReviewsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review/followed",
  validateSearch: (search: SearchRecord): PageSearch => ({
    page: numberParam(search.page),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/followed-reviews-page"),
    "FollowedReviewsPage"
  ),
})

export const userReviewsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/review/mine",
  validateSearch: (search: SearchRecord): UserCoursesSearch => ({
    page: numberParam(search.page),
  }),
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
  validateSearch: (search: SearchRecord): UserCoursesSearch => ({
    type: enumParam(search.type, ["enrolled", "followed", "ignored"] as const),
    page: numberParam(search.page),
    semester: rawStringParam(search.semester),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/user-courses-page"),
    "UserCoursesPage"
  ),
})

export const teachersRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/teacher",
  validateSearch: (search: SearchRecord): TeachersSearch => ({
    department: rawStringParam(search.department),
    title: rawStringParam(search.title),
    q: stringParam(search.q),
    page: numberParam(search.page),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/teachers-page"),
    "TeachersPage"
  ),
})

export const teacherDetailRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/teacher/$teacherID",
  validateSearch: (search: SearchRecord): TeacherDetailSearch => ({
    page: numberParam(search.page),
    order_by: enumParam(search.order_by, [
      "rating_score",
      "rating_count",
    ] as const),
  }),
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
  validateSearch: (search: SearchRecord): UserAdminSearch => ({
    tab: enumParam(search.tab, [
      "user",
      "admin",
      "system-api-key",
      "audit-log",
    ] as const),
    email: rawStringParam(search.email),
    page: numberParam(search.page),
    audit_page: numberParam(search.audit_page),
    audit_start_time: rawStringParam(search.audit_start_time),
    audit_end_time: rawStringParam(search.audit_end_time),
    audit_action: rawStringParam(search.audit_action),
    audit_actor_user_id: positiveIntParam(search.audit_actor_user_id, 0),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/admin/user-admin-page"),
    "UserAdminPage"
  ),
})

export const siteStatsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/admin/site-stat",
  validateSearch: (search: SearchRecord): SiteStatsSearch => ({
    start_date: rawStringParam(search.start_date),
    end_date: rawStringParam(search.end_date),
  }),
  component: lazyRouteComponent(
    () => import("@/pages/admin/site-stats-page"),
    "SiteStatsPage"
  ),
})

export const aboutRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/about",
  component: lazyRouteComponent(() => import("@/pages/about-page"), "AboutPage"),
})

export const faqRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/faq",
  component: lazyRouteComponent(() => import("@/pages/faq-page"), "FaqPage"),
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
