import {
  createRoute,
  lazyRouteComponent,
  type AnyRoute,
} from "@tanstack/react-router"

import { validateCanteenListSearch } from "./router-search"

export function createCanteenRoutes(appRoute: AnyRoute) {
  const canteenHomeRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/canteen",
    validateSearch: validateCanteenListSearch,
    component: lazyRouteComponent(
      () => import("@/pages/canteen-home-page"),
      "CanteenHomePage"
    ),
  })

  const canteenRankingsRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/canteen/rankings",
    validateSearch: validateCanteenListSearch,
    component: lazyRouteComponent(
      () => import("@/pages/canteen-rankings-page"),
      "CanteenRankingsPage"
    ),
  })

  const canteenStoreRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/canteen/stores/$storeID",
    component: lazyRouteComponent(
      () => import("@/pages/canteen-store-detail-page"),
      "CanteenStoreDetailPage"
    ),
  })

  const canteenReviewNewRoute = createRoute({
    getParentRoute: () => canteenStoreRoute,
    path: "review/new",
    component: lazyRouteComponent(
      () => import("@/pages/canteen-review-new-page"),
      "CanteenReviewNewPage"
    ),
  })

  const canteenDishDetailRoute = createRoute({
    getParentRoute: () => canteenStoreRoute,
    path: "dishes/$dishID",
    component: lazyRouteComponent(
      () => import("@/pages/canteen-dish-detail-page"),
      "CanteenDishDetailPage"
    ),
  })

  const canteenReviewEditRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/canteen/reviews/$reviewID/edit",
    component: lazyRouteComponent(
      () => import("@/pages/canteen-review-edit-page"),
      "CanteenReviewEditPage"
    ),
  })

  const canteenMyReviewsRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/canteen/reviews/mine",
    component: lazyRouteComponent(
      () => import("@/pages/canteen-my-reviews-page"),
      "CanteenMyReviewsPage"
    ),
  })

  const canteenContributionsRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/canteen/contributions",
    component: lazyRouteComponent(
      () => import("@/pages/canteen-contributions-page"),
      "CanteenContributionsPage"
    ),
  })

  const canteenAdminRoute = createRoute({
    getParentRoute: () => appRoute,
    path: "/admin/canteen",
    component: lazyRouteComponent(
      () => import("@/pages/admin/canteen-admin-page"),
      "CanteenAdminPage"
    ),
  })

  return {
    canteenHomeRoute,
    canteenRankingsRoute,
    canteenStoreRoute,
    canteenReviewNewRoute,
    canteenDishDetailRoute,
    canteenReviewEditRoute,
    canteenMyReviewsRoute,
    canteenContributionsRoute,
    canteenAdminRoute,
  }
}
