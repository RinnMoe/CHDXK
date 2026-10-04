import { copy } from "@/lib/copy"

export const ACCESS_KEYS = [
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

export const SHORTCUT_ROOT_SELECTOR = "[data-shortcut-root]"
export const SEARCH_TARGET_SELECTOR = "[data-shortcut-target='site-search']"

export type PageShortcut = {
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
    | "/admin/site-stat"
  requiresAuth?: boolean
  requiresAdmin?: boolean
}

export type ActionShortcut = {
  keys: string
  title: string
  description: string
}

export type AccessHint = {
  key: string
  label: string
  element: HTMLElement
  rect: DOMRect
}

export const pageShortcuts: PageShortcut[] = [
  {
    keys: "G H",
    title: copy("copy.user.components.shortcuts.shortcut-data.203c08e0d44a"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.6774a77de4da"
    ),
    path: "/",
  },
  {
    keys: "G C",
    title: copy("copy.user.components.shortcuts.shortcut-data.60db1afb4022"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.b5148d0abc51"
    ),
    path: "/course",
  },
  {
    keys: "G T",
    title: copy("copy.user.components.shortcuts.shortcut-data.a6654120cbec"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.3941c317f8fe"
    ),
    path: "/teacher",
  },
  {
    keys: "G R",
    title: copy("copy.user.components.shortcuts.shortcut-data.d2469bc43d2c"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.e5c0ab3568f5"
    ),
    path: "/review",
  },
  {
    keys: "G O",
    title: copy("copy.user.components.shortcuts.shortcut-data.d6e3b43107c5"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.8b3929779a11"
    ),
    path: "/course/hot",
  },
  {
    keys: "G F",
    title: copy("copy.user.components.shortcuts.shortcut-data.24a426113b37"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.1c5a05112a07"
    ),
    path: "/review/followed",
    requiresAuth: true,
  },
  {
    keys: "G M",
    title: copy("copy.user.components.shortcuts.shortcut-data.477d1ac374fc"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.d76ab02152f3"
    ),
    path: "/review/mine",
    requiresAuth: true,
  },
  {
    keys: "G U",
    title: copy("copy.user.components.shortcuts.shortcut-data.e80149e50579"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.b423bd0833b4"
    ),
    path: "/course/mine",
    requiresAuth: true,
  },
  {
    keys: "G P",
    title: copy("copy.user.components.shortcuts.shortcut-data.10939f548645"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.c590d930ddd9"
    ),
    path: "/point",
    requiresAuth: true,
  },
  {
    keys: "G A",
    title: copy("copy.user.components.shortcuts.shortcut-data.c9cd3687fc73"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.b28428e5fb41"
    ),
    path: "/admin/site-stat",
    requiresAdmin: true,
  },
]

export const actionShortcuts: ActionShortcut[] = [
  {
    keys: "/",
    title: copy("copy.user.components.shortcuts.shortcut-data.44ce7ae909bb"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.c062ec7f700d"
    ),
  },
  {
    keys: ".",
    title: copy("copy.user.components.shortcuts.shortcut-data.f8974e18945a"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.87006fb91401"
    ),
  },
  {
    keys: "?",
    title: copy("copy.user.components.shortcuts.shortcut-data.dc3b4a31f64f"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.c6871ff9fd78"
    ),
  },
  {
    keys: "Esc",
    title: copy("copy.user.components.shortcuts.shortcut-data.498e1d59b4d7"),
    description: copy(
      "copy.user.components.shortcuts.shortcut-data.9f469fdb28d4"
    ),
  },
]
