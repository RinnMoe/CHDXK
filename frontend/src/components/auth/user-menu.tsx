import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
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
      navigate("/")
    } finally {
      setIsLoggingOut(false)
    }
  }

  if (!user) {
    return (
      <div className="ml-auto flex gap-2">
        <Button variant="ghost" size="sm" asChild>
          <Link to="/login">登录</Link>
        </Button>
        <Button variant="default" size="sm" asChild>
          <Link to="/register">注册</Link>
        </Button>
      </div>
    )
  }

  const initials = user.username.slice(0, 2).toUpperCase()

  return (
    <div className="ml-auto">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="gap-2">
            <span className="flex size-6 items-center justify-center rounded-full bg-primary text-xs text-primary-foreground font-medium">
              {initials}
            </span>
            <span className="hidden sm:inline">{user.username}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem asChild>
            <Link to={`/reviews/mine`}>我的评价</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/courses/mine">我的课程</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/points">积分</Link>
          </DropdownMenuItem>
          {user.role === "admin" && (
            <DropdownMenuItem asChild>
              <Link to="/admin/site-stats">站点统计</Link>
            </DropdownMenuItem>
          )}
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleLogout} disabled={isLoggingOut}>
            退出登录
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
