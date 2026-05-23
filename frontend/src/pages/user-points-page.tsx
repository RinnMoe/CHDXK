import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { PageShell } from "@/components/layout/page-shell"
import { PointRecordList } from "@/components/point/point-record-list"
import { TransferForm } from "@/components/point/transfer-form"
import { useUserPoints } from "@/hooks/use-point"
import { useAuth } from "@/contexts/auth-context"

type View = "records" | "transfer"

export function UserPointsPage() {
  const { user } = useAuth()
  const [view, setView] = useState<View>("records")
  const [page] = useState(1)
  const [pageSize] = useState(20)
  const { data, isLoading } = useUserPoints(user?.id ?? 0, {
    page,
    page_size: pageSize,
  })

  if (!user) {
    return (
      <>
        <title>我的积分 - JCourse</title>
        <PageShell>
          <p className="py-12 text-center text-muted-foreground">请先登录</p>
        </PageShell>
      </>
    )
  }

  return (
    <>
      <title>我的积分 - JCourse</title>
      <PageShell>
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>我的积分</CardTitle>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Skeleton className="h-8 w-24" />
              ) : (
                <span className="text-3xl font-bold">{data?.total ?? 0}</span>
              )}
            </CardContent>
          </Card>

          <Tabs value={view} onValueChange={(value) => setView(value as View)}>
            <TabsList>
              <TabsTrigger value="records">积分记录</TabsTrigger>
              <TabsTrigger value="transfer">转账</TabsTrigger>
            </TabsList>

            <TabsContent value="records">
              {isLoading ? (
                <div className="space-y-3">
                  {[...Array(5)].map((_, i) => (
                    <Skeleton key={i} className="h-14 w-full" />
                  ))}
                </div>
              ) : (
                <PointRecordList records={data?.records.items ?? []} />
              )}
            </TabsContent>

            <TabsContent value="transfer">
              <Card>
                <CardContent className="pt-6">
                  <TransferForm onSuccess={() => setView("records")} />
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </PageShell>
    </>
  )
}
