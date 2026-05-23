import type { SiteDailyStatDTO } from "@/api/site-stats"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
  page: number
  pageSize: number
  total: number
}

export function DailyStatsTable({
  stats,
  page,
  pageSize,
  total,
}: DailyStatsTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>每日统计</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>日期</TableHead>
                <TableHead>用户总数</TableHead>
                <TableHead>点评总数</TableHead>
                <TableHead>活跃用户</TableHead>
                <TableHead>新增用户</TableHead>
                <TableHead>新增点评</TableHead>
                <TableHead>点评作者</TableHead>
                <TableHead>被点评课程</TableHead>
                <TableHead>点赞</TableHead>
                <TableHead>点踩</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {stats.map((s) => (
                <TableRow key={s.stat_date}>
                  <TableCell className="font-mono text-sm">
                    {s.stat_date}
                  </TableCell>
                  <TableCell>{s.total_user_count}</TableCell>
                  <TableCell>{s.total_review_count}</TableCell>
                  <TableCell>{s.active_user_count}</TableCell>
                  <TableCell>{s.new_user_count}</TableCell>
                  <TableCell>{s.new_review_count}</TableCell>
                  <TableCell>{s.review_author_count}</TableCell>
                  <TableCell>{s.reviewed_course_total}</TableCell>
                  <TableCell>{s.new_like_count}</TableCell>
                  <TableCell>{s.new_dislike_count}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        <p className="px-6 py-3 text-sm text-muted-foreground">
          第 {page} 页，共 {Math.ceil(total / pageSize)} 页，总计 {total} 条
        </p>
      </CardContent>
    </Card>
  )
}
