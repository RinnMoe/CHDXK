import { useState } from "react"

import { selectClass } from "@/components/admin/canteen/canteen-admin-shared"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import {
  useAdminCanteenReports,
  useResolveAdminCanteenReport,
} from "@/hooks/use-canteen-reports"

const reasonLabels: Record<string, string> = {
  spam: "垃圾广告",
  abuse: "辱骂攻击",
  false_info: "虚假信息",
  privacy: "隐私泄露",
  other: "其他",
}

export function CanteenReportPanel() {
  const [status, setStatus] = useState("open")
  const [error, setError] = useState<string | null>(null)
  const reports = useAdminCanteenReports(status, 1, 50)
  const resolve = useResolveAdminCanteenReport()

  async function handleResolve(
    reportID: number,
    nextStatus: "resolved" | "dismissed"
  ) {
    setError(null)
    try {
      await resolve.mutateAsync({ reportID, status: nextStatus })
    } catch (err) {
      setError(err instanceof Error ? err.message : "处理举报失败")
    }
  }

  return (
    <Card size="sm">
      <CardHeader>
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="text-base font-semibold">举报处理</h2>
          <select
            className={`${selectClass} ml-auto`}
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="open">待处理</option>
            <option value="resolved">已处理</option>
            <option value="dismissed">已驳回</option>
            <option value="">全部</option>
          </select>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {error && <p className="text-sm text-destructive">{error}</p>}
        {(reports.data?.items ?? []).length === 0 && (
          <p className="text-sm text-muted-foreground">没有举报记录。</p>
        )}
        {(reports.data?.items ?? []).map((report) => (
          <div key={report.id} className="space-y-2 rounded-md border p-3 text-sm">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p>
                #{report.id} · {report.subject_type} #{report.subject_id} ·{" "}
                {reasonLabels[report.reason_code] ?? report.reason_code}
              </p>
              <Badge variant="outline">{report.status}</Badge>
            </div>
            {report.detail && (
              <p className="text-muted-foreground">{report.detail}</p>
            )}
            <p className="text-xs text-muted-foreground">
              举报人 {report.reporter_user_id} ·{" "}
              {new Date(report.created_at).toLocaleString("zh-CN")}
            </p>
            {report.status === "open" && (
              <div className="flex gap-2">
                <Button
                  type="button"
                  size="xs"
                  onClick={() => void handleResolve(report.id, "resolved")}
                >
                  已处理
                </Button>
                <Button
                  type="button"
                  size="xs"
                  variant="outline"
                  onClick={() => void handleResolve(report.id, "dismissed")}
                >
                  驳回
                </Button>
              </div>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
