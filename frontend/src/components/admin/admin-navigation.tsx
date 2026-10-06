import { copy } from "@/lib/copy"
import { Button } from "@/components/ui/button"
import {
  adminSections,
  canAccessAdminTab,
  type AdminTab,
} from "./admin-navigation-model"

export function AdminNavigation({
  active,
  onChange,
  can,
}: {
  can: (permission: string) => boolean
  active: AdminTab
  onChange: (tab: AdminTab) => void
}) {
  return (
    <nav
      aria-label={copy(
        "copy.admin.components.admin.admin-navigation.41119ecad8a4"
      )}
      className="grid gap-4 rounded-lg border bg-card p-3 sm:grid-cols-2 lg:block lg:space-y-5"
    >
      {adminSections
        .map((section) => ({
          ...section,
          items: section.items.filter((item) =>
            canAccessAdminTab(item.key, can)
          ),
        }))
        .filter((section) => section.items.length > 0)
        .map((section) => (
          <section key={section.label} className="space-y-1">
            <h2 className="px-3 text-xs font-medium text-muted-foreground">
              {section.label}
            </h2>
            {section.items.map((item) => (
              <Button
                key={item.key}
                variant={active === item.key ? "secondary" : "ghost"}
                className="w-full justify-start"
                aria-current={active === item.key ? "page" : undefined}
                onClick={() => onChange(item.key)}
              >
                {item.label}
              </Button>
            ))}
          </section>
        ))}
    </nav>
  )
}
