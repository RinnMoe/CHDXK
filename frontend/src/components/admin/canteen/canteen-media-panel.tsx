import { useState } from "react"

import { canteenMediaURL } from "@/api/canteen-community"
import { selectClass } from "@/components/admin/canteen/canteen-admin-shared"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import {
  useAdminCanteenMedia,
  useReviewAdminCanteenMedia,
} from "@/hooks/use-canteen-media"

export function CanteenMediaPanel() {
  const [status, setStatus] = useState("pending")
  const [error, setError] = useState<string | null>(null)
  const media = useAdminCanteenMedia(status, 1, 50)
  const review = useReviewAdminCanteenMedia()

  async function handleReview(
    mediaID: number,
    nextStatus: "approved" | "rejected"
  ) {
    setError(null)
    try {
      await review.mutateAsync({ mediaID, status: nextStatus })
    } catch (err) {
      setError(err instanceof Error ? err.message : "审核图片失败")
    }
  }

  return (
    <Card size="sm">
      <CardHeader>
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="text-base font-semibold">实拍审核</h2>
          <select
            className={`${selectClass} ml-auto`}
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="pending">待审核</option>
            <option value="approved">已通过</option>
            <option value="rejected">未通过</option>
            <option value="">全部</option>
          </select>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {error && <p className="text-sm text-destructive">{error}</p>}
        {(media.data?.items ?? []).length === 0 && (
          <p className="text-sm text-muted-foreground">没有待审核图片。</p>
        )}
        {(media.data?.items ?? []).map((asset) => (
          <div
            key={asset.id}
            className="flex flex-wrap items-center gap-3 rounded-md border p-3 text-sm"
          >
            <img
              src={canteenMediaURL(asset.id)}
              alt="待审核实拍"
              className="h-20 w-20 rounded-md border object-cover"
            />
            <div className="space-y-1">
              <p>
                #{asset.id} · 用户 {asset.owner_user_id}
              </p>
              <p className="text-xs text-muted-foreground">
                {asset.width}×{asset.height} · {(asset.byte_size / 1024).toFixed(0)} KB
              </p>
              <Badge variant="outline">{asset.status}</Badge>
            </div>
            <div className="ml-auto flex gap-2">
              <Button
                type="button"
                size="xs"
                onClick={() => void handleReview(asset.id, "approved")}
              >
                通过
              </Button>
              <Button
                type="button"
                size="xs"
                variant="destructive"
                onClick={() => void handleReview(asset.id, "rejected")}
              >
                拒绝
              </Button>
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
