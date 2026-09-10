import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export function RegisterPage() {
  return (
    <>
      <PageTitle>注册</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <Card className="w-full max-w-sm">
            <CardHeader>
              <CardTitle>账号由主站管理</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm leading-6 text-muted-foreground">
                请在 res-hub 主站完成注册和账号设置，本站不保存独立账号。
              </p>
              <a
                href="/api/auth/sso/start?return_to=/"
                className="text-sm font-medium text-foreground underline underline-offset-4"
              >
                前往主站
              </a>
            </CardContent>
          </Card>
        </div>
      </PageShell>
    </>
  )
}
