import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { brand } from "@/config/brand"

const footerLinks = [
  {
    to: "/about",
    label: copy("copy.user.components.layout.site-footer.52d25a9e30ba"),
  },
  {
    to: "/faq",
    label: copy("copy.user.components.layout.site-footer.45a6d115fdfb"),
  },
]

export function SiteFooter() {
  const year = new Date().getFullYear()

  return (
    <footer className="border-t bg-background">
      <div className="mx-auto flex max-w-360 flex-col gap-3 px-4 py-6 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <p>
          &copy; {year} {brand.name}
          {copy("copy.user.components.layout.site-footer.7b45f5b81d1e")}
        </p>
        <nav
          className="flex gap-4"
          aria-label={copy(
            "copy.user.components.layout.site-footer.a512b3fc5b4b"
          )}
        >
          {footerLinks.map((link) => (
            <Link
              key={link.to}
              to={link.to}
              className="transition-colors hover:text-foreground"
            >
              {link.label}
            </Link>
          ))}
        </nav>
      </div>
    </footer>
  )
}
