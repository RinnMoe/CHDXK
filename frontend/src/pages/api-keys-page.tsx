import { copy } from "@/lib/copy"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export function ApiKeysPage() {
  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.api-keys-page.c08f17eb9e11")}
      </PageTitle>
      <PageShell>
        <Card className="mx-auto max-w-2xl">
          <CardHeader>
            <CardTitle>
              {copy("copy.user.pages.api-keys-page.b1b483588b65")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm leading-6 text-muted-foreground">
              {copy("copy.user.pages.api-keys-page.e07ad956e242")}
            </p>
          </CardContent>
        </Card>
      </PageShell>
    </>
  )
}
