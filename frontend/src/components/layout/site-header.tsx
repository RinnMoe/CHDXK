import { Link, useLocation } from "react-router-dom"
import { useState } from "react"
import { RiMenuLine } from "@remixicon/react"
import { UserMenu } from "@/components/auth/user-menu"
import { Button } from "@/components/ui/button"
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
    match: (p: string) => p.startsWith("/courses"),
  },
  {
    to: "/teachers",
    label: "教师",
    match: (p: string) => p.startsWith("/teachers"),
  },
  {
    to: "/reviews/latest",
    label: "点评",
    match: (p: string) => p.startsWith("/reviews"),
  },
  {
    to: "/courses/hot",
    label: "热门榜",
    match: (p: string) => p === "/courses/hot",
  },
]

export function SiteHeader() {
  const { pathname } = useLocation()
  const [open, setOpen] = useState(false)

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="mx-auto flex h-14 max-w-[1440px] items-center px-4">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <span className="text-lg">JCourse</span>
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
