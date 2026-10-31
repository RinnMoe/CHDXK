import { copy } from "@/lib/copy"
import { Link, useLocation, useNavigate } from "@tanstack/react-router"
import { useLayoutEffect, useRef, useState } from "react"
import { RiMenuLine, RiSearchLine } from "@remixicon/react"
import { UserMenu } from "@/components/auth/user-menu"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
import { useSystemSettings } from "@/hooks/use-system-settings"
import {
  Command,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { brand } from "@/config/brand"
import {
  Sheet,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetClose,
} from "@/components/ui/sheet"

type NavLinkItem = {
  to:
    | "/"
    | "/canteen"
    | "/course"
    | "/teacher"
    | "/review"
    | "/course/hot"
    | "/admin/user"
    | "/admin/canteen"
    | "/admin/site-stat"
  label: string
  match: (p: string) => boolean
}

const navItems: NavLinkItem[] = [
  {
    to: "/",
    label: copy("copy.user.components.layout.site-header.203c08e0d44a"),
    match: (p: string) => p === "/",
  },
  {
    to: "/course",
    label: copy("copy.user.components.layout.site-header.60db1afb4022"),
    match: (p: string) => p === "/course" || /^\/course\/\d+(?:\/|$)/.test(p),
  },
  {
    to: "/canteen",
    label: "食堂",
    match: (p: string) => p === "/canteen" || p.startsWith("/canteen/"),
  },
  {
    to: "/teacher",
    label: copy("copy.user.components.layout.site-header.a6654120cbec"),
    match: (p: string) => p.startsWith("/teacher"),
  },
  {
    to: "/review",
    label: copy("copy.user.components.layout.site-header.d2469bc43d2c"),
    match: (p: string) => p.startsWith("/review"),
  },
  {
    to: "/course/hot",
    label: copy("copy.user.components.layout.site-header.cdc394e2e036"),
    match: (p: string) => p === "/course/hot",
  },
]

const adminNavItems: NavLinkItem[] = [
  {
    to: "/admin/user",
    label: copy("copy.user.components.layout.site-header.bb6d995724f4"),
    match: (p: string) => p === "/admin/user" || p.startsWith("/admin/user/"),
  },
  {
    to: "/admin/canteen",
    label: "食堂管理",
    match: (p: string) =>
      p === "/admin/canteen" || p.startsWith("/admin/canteen/"),
  },
  {
    to: "/admin/site-stat",
    label: copy("copy.user.components.layout.site-header.89abe8c6ac13"),
    match: (p: string) =>
      p === "/admin/site-stat" || p.startsWith("/admin/site-stat/"),
  },
]

const searchTargets = [
  {
    label: copy("copy.user.components.layout.site-header.60db1afb4022"),
    path: "/course",
  },
  {
    label: copy("copy.user.components.layout.site-header.a6654120cbec"),
    path: "/teacher",
  },
  {
    label: copy("copy.user.components.layout.site-header.d2469bc43d2c"),
    path: "/review",
  },
] as const

function HeaderSearch() {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const keyword = query.trim()

  function go(path: (typeof searchTargets)[number]["path"]) {
    if (!keyword) return

    void navigate({
      to: path,
      search: { q: keyword },
    })
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="text-foreground"
          aria-expanded={open}
          aria-label={copy(
            "copy.user.components.layout.site-header.44ce7ae909bb"
          )}
          data-shortcut-target="site-search"
        >
          <RiSearchLine />
          <span className="sr-only">
            {copy("copy.user.components.layout.site-header.44ce7ae909bb")}
          </span>
        </Button>
      </PopoverTrigger>

      <PopoverContent
        align="end"
        sideOffset={8}
        className="w-[min(calc(100vw-2rem),20rem)]"
      >
        <Command shouldFilter={false} loop>
          <CommandInput
            autoFocus
            aria-label={copy(
              "copy.user.components.layout.site-header.c00a21fdf784"
            )}
            placeholder={copy(
              "copy.user.components.layout.site-header.c304ad38736d"
            )}
            value={query}
            onValueChange={setQuery}
          />

          {keyword && (
            <CommandList>
              {searchTargets.map((target) => (
                <CommandItem
                  key={target.path}
                  value={target.path}
                  onSelect={() => go(target.path)}
                >
                  {copy("copy.user.components.layout.site-header.859c3f6d9dff")}
                  {keyword}" {target.label}
                </CommandItem>
              ))}
            </CommandList>
          )}
        </Command>
      </PopoverContent>
    </Popover>
  )
}

export function SiteHeader() {
  const { pathname } = useLocation()
  const { user } = useAuth()
  const { data: publicSettings } = useSystemSettings()
  const canteenEnabled =
    publicSettings?.some(
      (item) => item.key === "canteen.enabled" && item.value === "true"
    ) ?? false
  const [open, setOpen] = useState(false)
  const navRef = useRef<HTMLElement>(null)
  const baseNavItems = canteenEnabled
    ? navItems
    : navItems.filter((item) => item.to !== "/canteen")
  const visibleNavItems: NavLinkItem[] = user?.is_admin()
    ? [...baseNavItems, ...adminNavItems]
    : baseNavItems
  const activeIndex = visibleNavItems.findIndex((item) => item.match(pathname))
  const [navIndicator, setNavIndicator] = useState({
    left: 0,
    width: 0,
    opacity: 0,
  })

  useLayoutEffect(() => {
    function updateIndicator() {
      const nav = navRef.current
      const activeLink =
        nav?.querySelectorAll<HTMLElement>("[data-nav-item]")[activeIndex]

      if (!activeLink) {
        setNavIndicator((current) => ({ ...current, opacity: 0 }))
        return
      }

      setNavIndicator({
        left: activeLink.offsetLeft,
        width: activeLink.offsetWidth,
        opacity: 1,
      })
    }

    updateIndicator()
    window.addEventListener("resize", updateIndicator)
    return () => window.removeEventListener("resize", updateIndicator)
  }, [activeIndex])

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
      <div className="mx-auto flex h-14 max-w-360 items-center px-2 sm:px-4">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <span className="text-lg">{brand.name}</span>
        </Link>

        {/* Desktop nav */}
        <nav
          ref={navRef}
          className="relative ml-8 hidden h-full gap-4 text-sm md:flex"
        >
          <span
            aria-hidden="true"
            className="pointer-events-none absolute bottom-0 h-0.5 rounded-full bg-primary transition-[left,width,opacity] duration-200 ease-out dark:bg-primary"
            style={navIndicator}
          />
          {visibleNavItems.map((item) => {
            const active = item.match(pathname)
            return (
              <Link
                key={item.to}
                to={item.to}
                data-nav-item
                className={
                  active
                    ? "flex items-center font-medium text-primary transition-colors dark:text-primary"
                    : "flex items-center text-foreground transition-colors hover:text-primary"
                }
              >
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <HeaderSearch />

          {/* Mobile menu trigger */}
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-foreground md:hidden"
              >
                <RiMenuLine />
                <span className="sr-only">
                  {copy("copy.user.components.layout.site-header.4ce4cafdd056")}
                </span>
              </Button>
            </SheetTrigger>
            <SheetContent side="right">
              <SheetHeader>
                <SheetTitle>
                  {copy("copy.user.components.layout.site-header.e72622fe470d")}
                </SheetTitle>
              </SheetHeader>
              <nav className="flex flex-col gap-1 px-4">
                {visibleNavItems.map((item) => {
                  const active = item.match(pathname)
                  return (
                    <SheetClose asChild key={item.to}>
                      <Link
                        to={item.to}
                        className={
                          active
                            ? "rounded-md bg-primary/10 px-3 py-2 font-medium text-primary dark:bg-primary/18 dark:text-primary"
                            : "rounded-md px-3 py-2 text-foreground transition-colors hover:bg-accent hover:text-primary"
                        }
                      >
                        {item.label}
                      </Link>
                    </SheetClose>
                  )
                })}
              </nav>
            </SheetContent>
          </Sheet>

          <UserMenu />
        </div>
      </div>
    </header>
  )
}
