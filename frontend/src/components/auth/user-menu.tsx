import { copy } from "@/lib/copy"
import { useState } from "react"
import { Link, useNavigate } from "@tanstack/react-router"
import { RiUserLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuth } from "@/contexts/auth-context"

export function UserMenu() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [isLoggingOut, setIsLoggingOut] = useState(false)

  async function handleLogout() {
    setIsLoggingOut(true)
    try {
      await logout()
      await navigate({ to: "/" })
    } finally {
      setIsLoggingOut(false)
    }
  }

  if (!user) {
    return (
      <div>
        <Button variant="ghost" size="sm" asChild>
          <Link to="/login">
            {copy("copy.user.components.auth.user-menu.1e2df9c3075a")}
          </Link>
        </Button>
      </div>
    )
  }

  return (
    <div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            className="text-foreground"
            aria-label={copy(
              "copy.user.components.auth.user-menu.c87328acafb2"
            )}
          >
            <span className="flex items-center justify-center">
              <RiUserLine aria-hidden="true" />
            </span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align="end"
          className="w-56 max-w-[calc(100vw-2rem)]"
        >
          <DropdownMenuLabel className="space-y-1.5 font-normal">
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 flex-1 text-sm font-medium wrap-anywhere text-foreground">
                {user.nickname?.trim() ||
                  user.bound_nickname?.trim() ||
                  user.username.trim() ||
                  copy("copy.user.components.auth.user-menu.0d0e1a86b3aa")}
              </p>
              <Badge variant="secondary" className="shrink-0 text-xs">
                {user.is_super_admin()
                  ? "超级管理员"
                  : user.is_admin()
                    ? copy("copy.user.components.auth.user-menu.e19796712f1c")
                    : copy("copy.user.components.auth.user-menu.f6a2faaac200")}
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground">
              {copy("copy.user.components.auth.user-menu.502e21b8a914")}
              {user.id}
            </p>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem asChild>
            <Link to="/review/mine">
              {copy("copy.user.components.auth.user-menu.477d1ac374fc")}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/course/mine">
              {copy("copy.user.components.auth.user-menu.e80149e50579")}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/point">
              {copy("copy.user.components.auth.user-menu.10939f548645")}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/settings">
              {copy("copy.user.components.auth.user-menu.df3d58c7d84b")}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleLogout} disabled={isLoggingOut}>
            {copy("copy.user.components.auth.user-menu.3ab8cc15939f")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
