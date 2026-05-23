import type { PointRecordDTO } from "@/api/point"
import { formatDateTime } from "@/lib/date"

const REASON_LABELS: Record<string, string> = {
  review_create: "发表点评",
  review_vote: "点评点赞",
  daily_login: "每日登录",
  transfer_in: "转账收入",
  transfer_out: "转账支出",
}

export function PointRecordList({ records }: { records: PointRecordDTO[] }) {
  if (records.length === 0) {
    return (
      <p className="py-6 text-center text-sm text-muted-foreground">
        暂无积分记录
      </p>
    )
  }
  return (
    <div className="divide-y rounded-md border">
      {records.map((r, i) => (
        <div
          key={i}
          className="flex items-start justify-between gap-4 px-4 py-3"
        >
          <div className="min-w-0 space-y-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold text-foreground">
                {REASON_LABELS[r.reason] ?? r.reason}
              </span>
              <span className="text-sm text-muted-foreground">
                {formatDateTime(r.created_at)}
              </span>
            </div>
            <p className="truncate text-sm font-normal text-muted-foreground">
              {r.description}
            </p>
          </div>
          <span
            className={
              r.amount >= 0
                ? "shrink-0 font-medium text-green-600"
                : "shrink-0 font-medium text-destructive"
            }
          >
            {r.amount >= 0 ? `+${r.amount}` : r.amount}
          </span>
        </div>
      ))}
    </div>
  )
}
