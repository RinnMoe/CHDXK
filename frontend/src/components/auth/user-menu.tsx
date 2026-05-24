import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
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
          <Button variant="ghost" size="icon" aria-label="打开用户菜单">
            <span className="flex size-6 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground">
              {initials}
            </span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
            用户 ID: {user.id}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem asChild>
            <Link to={`/review/mine`}>我的点评</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/course/mine">我的课程</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/point">积分</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/api-key">API Keys</Link>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleLogout} disabled={isLoggingOut}>
            退出登录
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
