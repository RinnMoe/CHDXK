import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export function ApiKeysPage() {
  return (
    <>
      <PageTitle>API Keys</PageTitle>
      <PageShell>
        <Card className="mx-auto max-w-2xl">
          <CardHeader>
            <CardTitle>接口凭证由主站管理</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm leading-6 text-muted-foreground">
              本站当前不提供个人 API 访问凭证。res-hub 与本站之间的集成凭证仅由服务端保管。
            </p>
          </CardContent>
        </Card>
      </PageShell>
    </>
  )
}
