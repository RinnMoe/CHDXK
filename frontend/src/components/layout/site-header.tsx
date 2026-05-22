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
  { to: "/courses", label: "课程", match: (p: string) => p.startsWith("/courses") },
  { to: "/teachers", label: "教师", match: (p: string) => p.startsWith("/teachers") },
  { to: "/reviews/latest", label: "点评", match: (p: string) => p.startsWith("/reviews") },
  { to: "/courses/hot", label: "热门榜", match: (p: string) => p === "/courses/hot" },
]

export function SiteHeader() {
  const { pathname } = useLocation()
  const [open, setOpen] = useState(false)

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center px-4 max-w-[1440px] mx-auto">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <span className="text-lg">JCourse</span>
        </Link>

        {/* Desktop nav */}
        <nav className="hidden md:flex ml-8 gap-4 text-sm h-full">
          {navItems.map((item) => {
            const active = item.match(pathname)
            return (
              <Link
                key={item.to}
                to={item.to}
                className={
                  active
                    ? "flex items-center text-primary font-medium border-b-2 border-primary"
                    : "flex items-center text-muted-foreground hover:text-foreground transition-colors"
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
                            ? "rounded-md px-3 py-2 text-primary font-medium bg-primary/10"
                            : "rounded-md px-3 py-2 text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
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
