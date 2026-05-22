import type { SiteDailyStatDTO } from "@/api/site-stats"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface StatItem {
  label: string
  value: number
}

function items(s: SiteDailyStatDTO): StatItem[] {
  return [
    { label: "活跃用户", value: s.active_user_count },
    { label: "新增用户", value: s.new_user_count },
    { label: "新增点评", value: s.new_review_count },
    { label: "点评作者", value: s.review_author_count },
    { label: "被点评课程总数", value: s.reviewed_course_total },
    { label: "新增点赞", value: s.new_like_count },
    { label: "新增点踩", value: s.new_dislike_count },
  ]
}

export function StatsCard({
  stat,
  title,
}: {
  stat: SiteDailyStatDTO
  title: string
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <p className="text-xs text-muted-foreground">{stat.stat_date}</p>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {items(stat).map((it) => (
            <div key={it.label} className="space-y-1">
              <p className="text-xs text-muted-foreground">{it.label}</p>
              <p className="text-xl font-semibold">{it.value}</p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
