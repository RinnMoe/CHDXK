import { useEffect, useMemo, useRef, useState } from "react"
import { useLocation, useNavigate } from "@tanstack/react-router"

import { useAuth } from "@/contexts/auth-context"
import {
  pageShortcuts,
  SEARCH_TARGET_SELECTOR,
  type AccessHint,
} from "./shortcut-data"
import {
  activateHint,
  collectAccessHints,
  isEditableTarget,
} from "./shortcut-dom"
import { AccessHintOverlay, ShortcutHelpDialog } from "./shortcut-help-dialog"

export function GlobalShortcuts() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user } = useAuth()
  const [helpOpen, setHelpOpen] = useState(false)
  const [accessRouteKey, setAccessRouteKey] = useState<string | null>(null)
  const [accessHints, setAccessHints] = useState<AccessHint[]>([])
  const pendingPrefixRef = useRef<string | null>(null)
  const prefixTimerRef = useRef<number | null>(null)
  const accessMode = accessRouteKey === location.href

  const availablePageShortcuts = useMemo(
    () =>
      pageShortcuts.filter((shortcut) => {
        if (shortcut.requiresAdmin) return user?.is_admin() ?? false
        if (shortcut.requiresAuth) return Boolean(user)
        return true
      }),
    [user]
  )

  useEffect(() => {
    if (!accessMode) {
      return undefined
    }

    const updateHints = () => {
      setAccessHints(collectAccessHints())
    }

    window.addEventListener("resize", updateHints)
    window.addEventListener("scroll", updateHints, true)

    return () => {
      window.removeEventListener("resize", updateHints)
      window.removeEventListener("scroll", updateHints, true)
    }
  }, [accessMode, location.href])

  useEffect(() => {
    return () => {
      if (prefixTimerRef.current !== null) {
        window.clearTimeout(prefixTimerRef.current)
      }
    }
  }, [])

  useEffect(() => {
    const clearPrefix = () => {
      pendingPrefixRef.current = null
      if (prefixTimerRef.current !== null) {
        window.clearTimeout(prefixTimerRef.current)
        prefixTimerRef.current = null
      }
    }

    const setPrefix = (prefix: string) => {
      pendingPrefixRef.current = prefix
      if (prefixTimerRef.current !== null) {
        window.clearTimeout(prefixTimerRef.current)
      }
      prefixTimerRef.current = window.setTimeout(clearPrefix, 1200)
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.repeat) return

      if (event.key === "Escape") {
        if (accessMode) {
          setAccessRouteKey(null)
          event.preventDefault()
        }
        clearPrefix()
        return
      }

      if (event.metaKey || event.ctrlKey || event.altKey) {
        return
      }

      if (isEditableTarget(event.target)) {
        return
      }

      const key = event.key.toLowerCase()

      if (event.key === "?" || (event.key === "/" && event.shiftKey)) {
        event.preventDefault()
        setHelpOpen((open) => !open)
        clearPrefix()
        return
      }

      if (helpOpen) {
        return
      }

      if (accessMode) {
        const hint = accessHints.find((item) => item.key === key)
        if (!hint) return

        event.preventDefault()
        activateHint(hint)
        setAccessRouteKey(null)
        return
      }

      if (key === "/") {
        const target = document.querySelector<HTMLElement>(
          SEARCH_TARGET_SELECTOR
        )
        if (target) {
          event.preventDefault()
          target.click()
        }
        clearPrefix()
        return
      }

      if (key === ".") {
        event.preventDefault()
        setAccessHints(collectAccessHints())
        setAccessRouteKey(location.href)
        clearPrefix()
        return
      }

      if (pendingPrefixRef.current === "g") {
        const shortcut = availablePageShortcuts.find(
          (item) => item.keys.toLowerCase() === `g ${key}`
        )
        clearPrefix()

        if (!shortcut) return

        event.preventDefault()
        void navigate({ to: shortcut.path })
        return
      }

      if (key === "g") {
        event.preventDefault()
        setPrefix("g")
      }
    }

    window.addEventListener("keydown", handleKeyDown)

    return () => {
      window.removeEventListener("keydown", handleKeyDown)
    }
  }, [
    accessHints,
    accessMode,
    availablePageShortcuts,
    helpOpen,
    location.href,
    navigate,
  ])

  return (
    <>
      {accessMode && <AccessHintOverlay hints={accessHints} />}
      <ShortcutHelpDialog
        open={helpOpen}
        onOpenChange={setHelpOpen}
        availablePageShortcuts={availablePageShortcuts}
      />
    </>
  )
}
