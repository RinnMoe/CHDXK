import type { PointRecordDTO } from "@/api/point"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatDateTime } from "@/lib/date"

export function PointRecordList({ records }: { records: PointRecordDTO[] }) {
  if (records.length === 0) {
    return (
      <p className="py-6 text-center text-sm text-muted-foreground">
        暂无积分记录
      </p>
    )
  }
  return (
    <div className="space-y-3">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>描述</TableHead>
            <TableHead className="w-24 text-right">积分</TableHead>
            <TableHead className="text-right">时间</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {records.map((r, i) => (
            <TableRow key={`${r.created_at}-${i}`}>
              <TableCell className="font-medium">{r.description}</TableCell>
              <TableCell
                className={
                  r.amount >= 0
                    ? "text-right font-medium text-green-600"
                    : "text-right font-medium text-destructive"
                }
              >
                {r.amount >= 0 ? `+${r.amount}` : r.amount}
              </TableCell>
              <TableCell className="text-right text-muted-foreground">
                {formatDateTime(r.created_at)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
