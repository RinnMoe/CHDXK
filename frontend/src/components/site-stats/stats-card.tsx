import { copy } from "@/lib/copy"
import type { SiteDailyStatDTO } from "@/api/site-stats"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatDateInputValue } from "@/lib/date"

interface StatItem {
  label: string
  value: number
}

function items(s: SiteDailyStatDTO): StatItem[] {
  return [
    {
      label: copy("copy.user.components.site-stats.stats-card.931f9c9d4123"),
      value: s.active_user_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.ebabc83b6830"),
      value: s.new_user_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.0b50583f022e"),
      value: s.new_review_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.a7d035e5fecb"),
      value: s.new_point_amount,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.c1320b7dd4f9"),
      value: s.review_author_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.e1a2345cdc47"),
      value: s.new_like_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.77d13c083dba"),
      value: s.new_dislike_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.545a676083c8"),
      value: s.total_user_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.c5ddc469e1d4"),
      value: s.total_review_count,
    },
    {
      label: copy("copy.user.components.site-stats.stats-card.86fa6f9d2237"),
      value: s.reviewed_course_total,
    },
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
        <p className="text-sm text-muted-foreground">
          {formatDateInputValue(stat.stat_date)}
        </p>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-5">
          {items(stat).map((it) => (
            <div key={it.label} className="space-y-1">
              <p className="text-sm text-muted-foreground">{it.label}</p>
              <p className="text-xl font-semibold">{it.value}</p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
