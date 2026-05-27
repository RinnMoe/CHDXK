import { useEffect, useLayoutEffect, useRef } from "react"
import { Outlet, useLocation, useNavigationType } from "react-router-dom"

export function RouteScrollRestoration() {
  const location = useLocation()
  const navigationType = useNavigationType()
  const locationKeyRef = useRef(location.key)
  const pathnameRef = useRef(location.pathname)
  const scrollPositionsRef = useRef(new Map<string, number>())

  useEffect(() => {
    window.history.scrollRestoration = "manual"

    return () => {
      window.history.scrollRestoration = "auto"
    }
  }, [])

  useLayoutEffect(() => {
    const previousLocationKey = locationKeyRef.current
    const previousPathname = pathnameRef.current

    scrollPositionsRef.current.set(previousLocationKey, window.scrollY)

    locationKeyRef.current = location.key
    pathnameRef.current = location.pathname

    if (previousPathname === location.pathname) return

    if (navigationType === "POP") {
      window.scrollTo({
        top: scrollPositionsRef.current.get(location.key) ?? 0,
        left: 0,
      })
      return
    }

    window.scrollTo({ top: 0, left: 0 })
  }, [location.key, location.pathname, navigationType])

  return null
}

export function RouterRoot() {
  return (
    <>
      <RouteScrollRestoration />
      <Outlet />
    </>
  )
}
