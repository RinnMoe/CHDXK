import { Link, useLocation, useNavigate } from "react-router-dom"
import { useState } from "react"
import { RiMenuLine, RiSearchLine } from "@remixicon/react"
import { UserMenu } from "@/components/auth/user-menu"
import { Button } from "@/components/ui/button"
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

const navItems = [
  { to: "/", label: "首页", match: (p: string) => p === "/" },
  {
    to: "/courses",
    label: "课程",
    match: (p: string) => p === "/courses" || /^\/courses\/\d+(?:\/|$)/.test(p),
  },
  {
    to: "/teachers",
    label: "教师",
    match: (p: string) => p.startsWith("/teachers"),
  },
  {
    to: "/reviews",
    label: "点评",
    match: (p: string) => p.startsWith("/reviews"),
  },
  {
    to: "/courses/hot",
    label: "热门",
    match: (p: string) => p === "/courses/hot",
  },
]

const searchTargets = [
  { label: "课程", path: "/courses" },
  { label: "教师", path: "/teachers" },
  { label: "点评", path: "/reviews" },
]

function HeaderSearch() {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const keyword = query.trim()

  function go(path: string) {
    if (!keyword) return

    navigate(`${path}?${new URLSearchParams({ q: keyword }).toString()}`)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          aria-expanded={open}
          aria-label="搜索"
          data-shortcut-target="site-search"
        >
          <RiSearchLine />
          <span className="sr-only">搜索</span>
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
            aria-label="搜索关键词"
            placeholder="搜索..."
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
                  搜索 "{keyword}" {target.label}
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
  const [open, setOpen] = useState(false)

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="mx-auto flex h-14 max-w-[1440px] items-center px-4">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <span className="text-lg">{brand.name}</span>
        </Link>

        {/* Desktop nav */}
        <nav className="ml-8 hidden h-full gap-4 text-sm md:flex">
          {navItems.map((item) => {
            const active = item.match(pathname)
            return (
              <Link
                key={item.to}
                to={item.to}
                className={
                  active
                    ? "flex items-center border-b-2 border-primary font-medium text-primary"
                    : "flex items-center text-muted-foreground transition-colors hover:text-foreground"
                }
              >
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <HeaderSearch />
          <UserMenu />

          {/* Mobile menu trigger */}
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon-sm" className="md:hidden">
                <RiMenuLine />
                <span className="sr-only">菜单</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="right">
              <SheetHeader>
                <SheetTitle>导航</SheetTitle>
              </SheetHeader>
              <nav className="flex flex-col gap-1 px-4">
                {navItems.map((item) => {
                  const active = item.match(pathname)
                  return (
                    <SheetClose asChild key={item.to}>
                      <Link
                        to={item.to}
                        className={
                          active
                            ? "rounded-md bg-primary/10 px-3 py-2 font-medium text-primary"
                            : "rounded-md px-3 py-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
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
        </div>
      </div>
    </header>
  )
}
