import { copy } from "@/lib/copy"
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
        {copy("copy.user.components.point.point-record-list.23069a877b5d")}
      </p>
    )
  }
  return (
    <div className="space-y-3">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>
              {copy(
                "copy.user.components.point.point-record-list.dc2ba467fc7a"
              )}
            </TableHead>
            <TableHead className="w-24 text-right">
              {copy(
                "copy.user.components.point.point-record-list.10939f548645"
              )}
            </TableHead>
            <TableHead className="text-right">
              {copy(
                "copy.user.components.point.point-record-list.8b6ff498515b"
              )}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {records.map((r, i) => (
            <TableRow key={`${r.created_at}-${i}`}>
              <TableCell className="font-medium">{r.description}</TableCell>
              <TableCell
                className={
                  r.amount >= 0
                    ? "text-right font-medium text-success"
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
