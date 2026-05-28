import { useEffect, useMemo, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { useLocation, useNavigate } from "@tanstack/react-router"

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Kbd, KbdGroup } from "@/components/ui/kbd"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useAuth } from "@/contexts/auth-context"
import { cn } from "@/lib/utils"

const ACCESS_KEYS = [
  "1",
  "2",
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
  "a",
  "s",
  "f",
  "j",
  "k",
  "l",
  "q",
  "w",
  "e",
  "r",
  "t",
  "y",
  "u",
  "i",
  "o",
  "p",
  "z",
  "x",
  "c",
  "v",
  "b",
  "n",
  "m",
]

const SHORTCUT_ROOT_SELECTOR = "[data-shortcut-root]"
const SEARCH_TARGET_SELECTOR = "[data-shortcut-target='site-search']"
const ACCESS_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "input:not([type='hidden']):not([disabled])",
  "textarea:not([disabled])",
  "select:not([disabled])",
  "[role='button']:not([aria-disabled='true'])",
  "[role='menuitem']:not([aria-disabled='true'])",
].join(",")

type PageShortcut = {
  keys: string
  title: string
  description: string
  path:
    | "/"
    | "/course"
    | "/teacher"
    | "/review"
    | "/course/hot"
    | "/review/followed"
    | "/review/mine"
    | "/course/mine"
    | "/point"
    | "/api-key"
    | "/admin/site-stat"
  requiresAuth?: boolean
  requiresAdmin?: boolean
}

type ActionShortcut = {
  keys: string
  title: string
  description: string
}

type AccessHint = {
  key: string
  label: string
  element: HTMLElement
  rect: DOMRect
}

const pageShortcuts: PageShortcut[] = [
  { keys: "G H", title: "首页", description: "回到首页", path: "/" },
  {
    keys: "G C",
    title: "课程",
    description: "浏览和搜索课程",
    path: "/course",
  },
  {
    keys: "G T",
    title: "教师",
    description: "浏览和搜索教师",
    path: "/teacher",
  },
  {
    keys: "G R",
    title: "点评",
    description: "查看公开点评",
    path: "/review",
  },
  {
    keys: "G O",
    title: "热门课程",
    description: "查看热门课程",
    path: "/course/hot",
  },
  {
    keys: "G F",
    title: "关注点评",
    description: "查看关注课程的点评",
    path: "/review/followed",
    requiresAuth: true,
  },
  {
    keys: "G M",
    title: "我的点评",
    description: "查看我发布的点评",
    path: "/review/mine",
    requiresAuth: true,
  },
  {
    keys: "G U",
    title: "我的课程",
    description: "查看我关注的课程",
    path: "/course/mine",
    requiresAuth: true,
  },
  {
    keys: "G P",
    title: "积分",
    description: "查看积分流水",
    path: "/point",
    requiresAuth: true,
  },
  {
    keys: "G K",
    title: "API Keys",
    description: "管理 API Key",
    path: "/api-key",
    requiresAuth: true,
  },
  {
    keys: "G A",
    title: "站点统计",
    description: "查看管理统计",
    path: "/admin/site-stat",
    requiresAdmin: true,
  },
]

const actionShortcuts: ActionShortcut[] = [
  {
    keys: "/",
    title: "搜索",
    description: "打开顶部搜索框",
  },
  {
    keys: ".",
    title: "页面控件",
    description: "给当前可见按钮、链接和输入框显示编号",
  },
  {
    keys: "?",
    title: "快捷键提示",
    description: "打开或关闭此提示",
  },
  {
    keys: "Esc",
    title: "退出",
    description: "关闭提示或退出页面控件模式",
  },
]

function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) {
    return false
  }

  if (target.isContentEditable) {
    return true
  }

  return Boolean(target.closest("input, textarea, select, [contenteditable]"))
}

function isVisibleElement(element: HTMLElement) {
  if (element.closest(SHORTCUT_ROOT_SELECTOR)) {
    return false
  }

  if (element.getAttribute("aria-hidden") === "true") {
    return false
  }

  const style = window.getComputedStyle(element)
  if (style.visibility === "hidden" || style.display === "none") {
    return false
  }

  const rect = element.getBoundingClientRect()
  return (
    rect.width > 0 &&
    rect.height > 0 &&
    rect.bottom >= 0 &&
    rect.right >= 0 &&
    rect.top <= window.innerHeight &&
    rect.left <= window.innerWidth
  )
}

function getElementLabel(element: HTMLElement) {
  const ariaLabel = element.getAttribute("aria-label")?.trim()
  if (ariaLabel) return ariaLabel

  const title = element.getAttribute("title")?.trim()
  if (title) return title

  const placeholder = element.getAttribute("placeholder")?.trim()
  if (placeholder) return placeholder

  const text = element.textContent?.replace(/\s+/g, " ").trim()
  if (text) return text.slice(0, 32)

  if (element instanceof HTMLInputElement) {
    return element.name || "输入框"
  }

  return "控件"
}

function collectAccessHints() {
  const candidates = Array.from(
    document.querySelectorAll<HTMLElement>(ACCESS_SELECTOR)
  )

  return candidates
    .filter(isVisibleElement)
    .slice(0, ACCESS_KEYS.length)
    .map((element, index) => ({
      key: ACCESS_KEYS[index] ?? "",
      label: getElementLabel(element),
      element,
      rect: element.getBoundingClientRect(),
    }))
}

function isFocusableControl(element: HTMLElement) {
  return (
    element instanceof HTMLInputElement ||
    element instanceof HTMLTextAreaElement ||
    element instanceof HTMLSelectElement
  )
}

function activateHint(hint: AccessHint) {
  if (isFocusableControl(hint.element)) {
    hint.element.focus()
    if (
      hint.element instanceof HTMLInputElement ||
      hint.element instanceof HTMLTextAreaElement
    ) {
      hint.element.select()
    }
    return
  }

  hint.element.click()
}

function ShortcutKeys({ keys }: { keys: string }) {
  return (
    <KbdGroup className="shrink-0">
      {keys.split(" ").map((key) => (
        <Kbd key={key} className="min-w-6 text-foreground">
          {key}
        </Kbd>
      ))}
    </KbdGroup>
  )
}

function ShortcutRow({
  keys,
  title,
  description,
  disabled,
}: {
  keys: string
  title: string
  description: string
  disabled?: boolean
}) {
  return (
    <div
      className={cn(
        "grid grid-cols-[4.75rem_minmax(0,1fr)] items-start gap-3 rounded-md px-2 py-2",
        disabled && "opacity-45"
      )}
    >
      <ShortcutKeys keys={keys} />
      <div className="min-w-0">
        <div className="font-medium text-foreground">{title}</div>
        <div className="text-xs leading-5 text-muted-foreground">
          {description}
        </div>
      </div>
    </div>
  )
}

function AccessHintOverlay({ hints }: { hints: AccessHint[] }) {
  return createPortal(
    <div
      data-shortcut-root
      className="pointer-events-none fixed inset-0 z-[70]"
      aria-hidden="true"
    >
      <div className="absolute top-16 left-1/2 -translate-x-1/2 rounded-md border bg-popover px-3 py-2 text-xs text-popover-foreground shadow-lg">
        输入标记可点击或聚焦控件，Esc 退出
      </div>
      {hints.map((hint) => (
        <Kbd
          key={`${hint.key}-${hint.label}-${hint.rect.left}-${hint.rect.top}`}
          className="absolute border border-primary bg-primary px-1.5 font-mono font-bold text-primary-foreground shadow-lg"
          style={{
            top: Math.max(4, hint.rect.top - 8),
            left: Math.max(4, hint.rect.left - 8),
          }}
        >
          {hint.key.toUpperCase()}
        </Kbd>
      ))}
    </div>,
    document.body
  )
}

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
      <Dialog open={helpOpen} onOpenChange={setHelpOpen}>
        <DialogContent
          data-shortcut-root
          className="max-h-[min(42rem,calc(100svh-2rem))] grid-rows-[auto_minmax(0,1fr)] overflow-hidden sm:max-w-2xl"
        >
          <DialogHeader>
            <DialogTitle>快捷键</DialogTitle>
            <DialogDescription>
              页面跳转可先按 G 再按目标键；输入框内不会触发全局快捷键。
            </DialogDescription>
          </DialogHeader>

          <ScrollArea className="-mx-2 min-h-0 pr-3">
            <div className="space-y-5 px-2 pb-1">
              <section>
                <h3 className="mb-2 text-sm font-medium">页面</h3>
                <div className="space-y-1">
                  {pageShortcuts.map((shortcut) => {
                    const disabled = !availablePageShortcuts.includes(shortcut)
                    return (
                      <ShortcutRow
                        key={shortcut.path}
                        keys={shortcut.keys}
                        title={shortcut.title}
                        description={
                          disabled
                            ? shortcut.requiresAdmin
                              ? "需要管理员权限"
                              : "需要登录"
                            : shortcut.description
                        }
                        disabled={disabled}
                      />
                    )
                  })}
                </div>
              </section>

              <section>
                <h3 className="mb-2 text-sm font-medium">操作</h3>
                <div className="space-y-1">
                  {actionShortcuts.map((shortcut) => (
                    <ShortcutRow
                      key={shortcut.keys}
                      keys={shortcut.keys}
                      title={shortcut.title}
                      description={shortcut.description}
                    />
                  ))}
                </div>
              </section>
            </div>
          </ScrollArea>
        </DialogContent>
      </Dialog>
    </>
  )
}
