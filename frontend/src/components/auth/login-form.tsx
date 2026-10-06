import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { getRouteApi } from "@tanstack/react-router"
import {
  getOIDCStartURL,
  getAuthOptions,
  passwordLogin,
  registerAccount,
} from "@/api/auth"
import { getSafeRedirectPath } from "@/lib/auth-redirect"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

const routeApi = getRouteApi("/public/login")
export function LoginForm() {
  const search = routeApi.useSearch()
  const returnPath = getSafeRedirectPath(search.redirect)
  const { data: options } = useQuery({
    queryKey: ["auth", "options"],
    queryFn: getAuthOptions,
    staleTime: 0,
  })
  const [mode, setMode] = useState<"closed" | "login" | "register">("closed")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)
  const registering = mode === "register" && !!options?.registration_enabled
  async function submit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setBusy(true)
    setError("")
    const data = new FormData(event.currentTarget)
    try {
      const username = (data.get("username") as string | null) ?? ""
      const password = (data.get("password") as string | null) ?? ""
      const user = registering
        ? await registerAccount(
            username,
            (data.get("nickname") as string | null) ?? "",
            password
          )
        : await passwordLogin(username, password)
      window.location.replace(
        user.must_change_password ? "/change-password" : returnPath
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败，请稍后重试")
      setBusy(false)
    }
  }
  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>登录</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        <p className="text-sm leading-6 text-muted-foreground">
          使用资料站账号，继续查看课程、分享点评。
        </p>
        {search.error && (
          <p role="alert" className="text-sm text-destructive">
            资料站登录未完成，请重试，或使用本站账号登录。
          </p>
        )}
        <Button asChild className="w-full">
          <a href={getOIDCStartURL(returnPath)}>通过资料站登录</a>
        </Button>
        {mode === "closed" ? (
          <button
            type="button"
            className="block w-full text-center text-xs text-muted-foreground hover:text-foreground"
            onClick={() => setMode("login")}
          >
            使用本站账号登录
          </button>
        ) : (
          <div className="space-y-4 border-t pt-4">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-medium">
                {registering ? "注册本站账号" : "本站账号登录"}
              </h2>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setMode("closed")}
              >
                收起
              </Button>
            </div>
            <form
              onSubmit={(event) => void submit(event)}
              className="space-y-4"
            >
              <div className="space-y-2">
                <Label htmlFor="local-username">账号</Label>
                <Input
                  id="local-username"
                  name="username"
                  autoComplete="username"
                  required
                  minLength={3}
                  maxLength={32}
                />
              </div>
              {registering && (
                <div className="space-y-2">
                  <Label htmlFor="local-nickname">昵称</Label>
                  <Input
                    id="local-nickname"
                    name="nickname"
                    required
                    maxLength={64}
                  />
                </div>
              )}
              <div className="space-y-2">
                <Label htmlFor="local-password">密码</Label>
                <Input
                  id="local-password"
                  name="password"
                  type="password"
                  autoComplete={
                    registering ? "new-password" : "current-password"
                  }
                  required
                  minLength={registering ? 8 : undefined}
                />
              </div>
              {registering && (
                <p className="text-xs text-muted-foreground">
                  账号须为 3–32 位小写字母、数字、点、横线或下划线；密码至少 8
                  个字符。
                </p>
              )}
              {error && (
                <p role="alert" className="text-sm text-destructive">
                  {error}
                </p>
              )}
              <Button disabled={busy} className="w-full" type="submit">
                {busy ? "请稍候…" : registering ? "注册并登录" : "登录"}
              </Button>
            </form>
            {options?.registration_enabled && (
              <Button
                type="button"
                variant="link"
                size="sm"
                className="w-full"
                disabled={busy}
                onClick={() => {
                  setMode(registering ? "login" : "register")
                  setError("")
                }}
              >
                {registering ? "已有本站账号，返回登录" : "注册本站账号"}
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
