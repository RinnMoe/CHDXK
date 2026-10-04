import { copy } from "@/lib/copy"
import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { PointRecordList } from "@/components/point/point-record-list"
import { useUserPoints } from "@/hooks/use-point"
import { useAuth } from "@/contexts/auth-context"
import { getErrorMessage } from "@/components/admin/admin-utils"

export function UserPointsPage() {
  const { user } = useAuth()
  const [page] = useState(1)
  const [pageSize] = useState(20)
  const { data, isLoading, isError, error } = useUserPoints(user?.id ?? 0, {
    page,
    page_size: pageSize,
  })

  if (!user) {
    return (
      <>
        <PageTitle>
          {copy("copy.user.pages.user-points-page.e8c4f71b3601")}
        </PageTitle>
        <PageShell>
          <p className="py-12 text-center text-muted-foreground">
            {copy("copy.user.pages.user-points-page.48144f96c150")}
          </p>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.user-points-page.e8c4f71b3601")}
      </PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold">
              {copy("copy.user.pages.user-points-page.e8c4f71b3601")}
            </h1>
            <p className="text-sm text-muted-foreground">
              {copy("copy.user.pages.user-points-page.0a6a74474ce7")}
            </p>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>
                {copy("copy.user.pages.user-points-page.e8c4f71b3601")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Skeleton className="h-8 w-24" />
              ) : (
                <span className="text-3xl font-bold">{data?.total ?? 0}</span>
              )}
            </CardContent>
          </Card>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">
              {copy("copy.user.pages.user-points-page.ed8745d001b1")}
            </h2>
            {isLoading ? (
              <div className="space-y-3">
                {[...Array(5)].map((_, i) => (
                  <Skeleton key={i} className="h-14 w-full" />
                ))}
              </div>
            ) : isError ? (
              <p className="text-sm text-destructive" role="alert">
                {getErrorMessage(error)}
              </p>
            ) : (
              <PointRecordList records={data?.records.items ?? []} />
            )}
          </section>
        </div>
      </PageShell>
    </>
  )
}
