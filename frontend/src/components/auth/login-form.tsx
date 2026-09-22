import { useEffect } from "react"
import { getRouteApi } from "@tanstack/react-router"
import { getOIDCStartURL } from "@/api/auth"
import { getSafeRedirectPath } from "@/lib/auth-redirect"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

const routeApi = getRouteApi("/public/login")

export function LoginForm() {
  const search = routeApi.useSearch()
  const returnPath = getSafeRedirectPath(search.redirect)
  const oidcURL = getOIDCStartURL(returnPath)

  useEffect(() => {
    if (!search.error) window.location.replace(oidcURL)
  }, [search.error, oidcURL])

  const errorMessage =
    search.error === "oidc_denied"
      ? "你取消了主站登录。"
      : search.error === "oidc_unavailable"
        ? "主站暂时不可用，请稍后重试。"
        : search.error === "oidc_failed"
          ? "主站登录未能完成，请重试。"
          : null

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>使用主站登录</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm leading-6 text-muted-foreground">
          本站复用主站的账号、角色、登录状态和积分。登录后会自动返回本站。
        </p>
        {errorMessage ? (
          <p className="text-sm text-destructive" role="alert">
            {errorMessage}
          </p>
        ) : null}
        <Button asChild className="w-full">
          <a href={oidcURL}>前往主站登录</a>
        </Button>
      </CardContent>
    </Card>
  )
}
