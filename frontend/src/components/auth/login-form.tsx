import { copy } from "@/lib/copy"
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
      ? copy("copy.user.components.auth.login-form.01c59316ce41")
      : search.error === "oidc_unavailable"
        ? copy("copy.user.components.auth.login-form.35e0a6baf5e8")
        : search.error === "oidc_failed"
          ? copy("copy.user.components.auth.login-form.ed2dc40e25ea")
          : null

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>
          {copy("copy.user.components.auth.login-form.7a5c4255247b")}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm leading-6 text-muted-foreground">
          {copy("copy.user.components.auth.login-form.66c5fcb2ee11")}
        </p>
        {errorMessage ? (
          <p className="text-sm text-destructive" role="alert">
            {errorMessage}
          </p>
        ) : null}
        <Button asChild className="w-full">
          <a href={oidcURL}>
            {copy("copy.user.components.auth.login-form.f44a317e658b")}
          </a>
        </Button>
      </CardContent>
    </Card>
  )
}
