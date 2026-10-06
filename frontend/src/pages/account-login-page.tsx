import { copy } from "@/lib/copy"

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { getLoginAccounts, selectLoginAccount } from "@/api/auth"
import { getSafeRedirectPath } from "@/lib/auth-redirect"
import { Button } from "@/components/ui/button"
import { PageShell } from "@/components/layout/page-shell"

export function LoginAccountsPage() {
  const { data, isPending, error } = useQuery({
    queryKey: ["login-accounts"],
    queryFn: getLoginAccounts,
    retry: false,
  })
  const [busy, setBusy] = useState(false)
  const [failure, setFailure] = useState("")
  async function choose(id: number) {
    setBusy(true)
    try {
      const selected = await selectLoginAccount(id)
      window.location.replace(
        selected.must_change_password
          ? "/change-password"
          : getSafeRedirectPath(
              new URLSearchParams(window.location.search).get("redirect") ??
                undefined
            )
      )
    } catch {
      setFailure(copy("copy.user.pages.account-login-page.1e57edd9c951"))
      setBusy(false)
    }
  }
  return (
    <PageShell>
      <div className="mx-auto max-w-lg space-y-6 py-12">
        <h1 className="text-2xl font-semibold">
          {copy("copy.user.pages.account-login-page.803789846bba")}
        </h1>
        <p className="text-muted-foreground">
          {copy("copy.user.pages.account-login-page.df2d3333e46e")}
        </p>
        {isPending && (
          <p>{copy("copy.user.pages.account-login-page.4927a53bcc88")}</p>
        )}
        {(error || failure) && (
          <p role="alert">
            {failure || copy("copy.user.pages.account-login-page.bd0bb959fbb4")}
          </p>
        )}
        {data?.map((account) => {
          const banned =
            !!account.suspended_at &&
            (!account.suspend_till ||
              new Date(account.suspend_till) > new Date())
          return (
            <Button
              key={account.id}
              data-account-id={account.id}
              variant="outline"
              className="h-auto w-full justify-between p-4"
              disabled={busy || banned}
              onClick={() => void choose(account.id)}
            >
              <span className="text-left">
                <span className="block">
                  {account.nickname ||
                    account.bound_nickname ||
                    account.username}
                </span>
                <span className="mt-1 block text-xs text-muted-foreground">
                  {account.is_bootstrap
                    ? copy("copy.user.pages.account-login-page.cbe255d58fc5")
                    : account.role === "user"
                      ? copy("copy.user.pages.account-login-page.f6a2faaac200")
                      : copy(
                          "copy.user.pages.account-login-page.e19796712f1c"
                        )}{" "}
                  {copy("copy.user.pages.account-login-page.bb9e59ffebc4")}
                  {account.id}
                </span>
              </span>
              <span>
                {banned
                  ? "本站账号已封禁"
                  : copy("copy.user.pages.account-login-page.1e2df9c3075a")}
              </span>
            </Button>
          )
        })}
      </div>
    </PageShell>
  )
}
