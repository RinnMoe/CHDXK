import {
  createRoute,
  lazyRouteComponent,
  type AnyRoute,
} from "@tanstack/react-router"

export function createPublicContentRoutes(publicRoute: AnyRoute) {
  const aboutRoute = createRoute({
    getParentRoute: () => publicRoute,
    path: "/about",
    component: lazyRouteComponent(
      () => import("@/pages/about-page"),
      "AboutPage"
    ),
  })

  const faqRoute = createRoute({
    getParentRoute: () => publicRoute,
    path: "/faq",
    component: lazyRouteComponent(() => import("@/pages/faq-page"), "FaqPage"),
  })

  const sitePageRoute = createRoute({
    getParentRoute: () => publicRoute,
    path: "/p/$slug",
    component: lazyRouteComponent(
      () => import("@/pages/site-page"),
      "SitePage"
    ),
  })

  return { aboutRoute, faqRoute, sitePageRoute }
}
