import { copy } from "@/lib/copy"

export const adminSections = [
  {
    label: copy("copy.admin.components.admin.admin-navigation.fea448341e76"),
    items: [
      {
        key: "course-data",
        label: copy(
          "copy.admin.components.admin.admin-navigation.014df7b8215a"
        ),
      },
    ],
  },
  {
    label: copy("copy.admin.components.admin.admin-navigation.70a3c398dfd4"),
    items: [
      {
        key: "announcement",
        label: copy(
          "copy.admin.components.admin.admin-navigation.838aeb39d693"
        ),
      },
      {
        key: "content-pages",
        label: copy(
          "copy.admin.components.admin.admin-navigation.f52222e9c24c"
        ),
      },
    ],
  },
  {
    label: copy("copy.admin.components.admin.admin-navigation.11258110296f"),
    items: [
      {
        key: "system-settings",
        label: copy(
          "copy.admin.components.admin.admin-navigation.d9671a0813d2"
        ),
      },
      {
        key: "appearance",
        label: copy(
          "copy.admin.components.admin.admin-navigation-model.6b6b25f98162"
        ),
      },
      {
        key: "site-copy",
        label: copy(
          "copy.admin.components.admin.admin-navigation.12abdae1e423"
        ),
      },
    ],
  },
  {
    label: copy("copy.admin.components.admin.admin-navigation.f586ac06afb9"),
    items: [
      {
        key: "system-api-key",
        label: copy(
          "copy.admin.components.admin.admin-navigation.491c5e8e43a5"
        ),
      },
      {
        key: "audit-log",
        label: copy(
          "copy.admin.components.admin.admin-navigation.b2db0d03bb7d"
        ),
      },
    ],
  },
] as const

export type AdminTab = (typeof adminSections)[number]["items"][number]["key"]
export function getAdminTab(value?: string | null): AdminTab {
  return adminSections.some((section) =>
    section.items.some((item) => item.key === value)
  )
    ? (value as AdminTab)
    : "course-data"
}
