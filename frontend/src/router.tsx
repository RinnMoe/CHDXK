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
  const user = await context.queryClient.fetchQuery(authMeQueryOptions())
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
  const user = await context.queryClient.fetchQuery(authMeQueryOptions())
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
  component: lazyRouteComponent(
    () => import("@/pages/login-page"),
    "LoginPage"
  ),
})

const bootstrapLoginRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/bootstrap-login",
  component: lazyRouteComponent(
    () => import("@/pages/account-login-page"),
    "BootstrapLoginPage"
  ),
})
const loginAccountsRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/login/accounts",
  component: lazyRouteComponent(
    () => import("@/pages/account-login-page"),
    "LoginAccountsPage"
  ),
})

export const coursesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/course",
  validateSearch: validateCourseListSearch,
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
  validateSearch: validateCourseDetailSearch,
  component: lazyRouteComponent(
    () => import("@/pages/course-detail-page"),
    "CourseDetailPage"
  ),
})

export const newReviewRoute = createRoute({
  getParentRoute: () => courseDetailRoute,
  path: "/review/new",
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
  getParentRoute: () => reviewDetailRoute,
  path: "/edit",
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
  getParentRoute: () => publicRoute,
  path: "/about",
  component: lazyRouteComponent(
    () => import("@/pages/about-page"),
    "AboutPage"
  ),
})

export const faqRoute = createRoute({
  getParentRoute: () => publicRoute,
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

export const sitePageRoute = createRoute({
  getParentRoute: () => publicRoute,
  path: "/p/$slug",
  component: lazyRouteComponent(() => import("@/pages/site-page"), "SitePage"),
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
    bootstrapLoginRoute,
    loginAccountsRoute,
    loginRoute,
    aboutRoute,
    faqRoute,
    sitePageRoute,
  ]),
  appRoute.addChildren([
    homeRoute,
    coursesRoute,
    hotCoursesRoute,
    courseDetailRoute.addChildren([newReviewRoute]),
    reviewsRoute,
    followedReviewsRoute,
    userReviewsRoute,
    userCoursesRoute,
    reviewDetailRoute.addChildren([editReviewRoute]),
    teachersRoute,
    teacherDetailRoute,
    userPointsRoute,
    apiKeysRoute,
    userSettingsRoute,
    userAdminRoute,
    siteStatsRoute,
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
