import { copy } from "@/lib/copy"
import type { SiteDailyStatDTO } from "@/api/site-stats"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatDateInputValue } from "@/lib/date"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface DailyStatsTableProps {
  stats: SiteDailyStatDTO[]
}

export function DailyStatsTable({ stats }: DailyStatsTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {copy(
            "copy.user.components.site-stats.daily-stats-table.9f34653227fc"
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.70d0c1b33626"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.931f9c9d4123"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.ebabc83b6830"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.0b50583f022e"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.a7d035e5fecb"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.c1320b7dd4f9"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.e1a2345cdc47"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.77d13c083dba"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.545a676083c8"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.c5ddc469e1d4"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.user.components.site-stats.daily-stats-table.51f48f6d74a7"
                  )}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {stats.map((s) => (
                <TableRow key={s.stat_date}>
                  <TableCell className="font-mono text-sm">
                    {formatDateInputValue(s.stat_date)}
                  </TableCell>
                  <TableCell>{s.active_user_count}</TableCell>
                  <TableCell>{s.new_user_count}</TableCell>
                  <TableCell>{s.new_review_count}</TableCell>
                  <TableCell>{s.new_point_amount}</TableCell>
                  <TableCell>{s.review_author_count}</TableCell>
                  <TableCell>{s.new_like_count}</TableCell>
                  <TableCell>{s.new_dislike_count}</TableCell>
                  <TableCell>{s.total_user_count}</TableCell>
                  <TableCell>{s.total_review_count}</TableCell>
                  <TableCell>{s.reviewed_course_total}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
