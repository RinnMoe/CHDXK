import { copy } from "@/lib/copy"
import type { SiteDailyStatDTO } from "@/api/site-stats"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatDateInputValue } from "@/lib/date"
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts"

interface DailyStatsChartProps {
  stats: SiteDailyStatDTO[]
}

type MetricKey =
  | "active_user_count"
  | "new_user_count"
  | "new_review_count"
  | "new_point_amount"
  | "review_author_count"
  | "total_user_count"
  | "total_review_count"
  | "reviewed_course_total"
  | "new_like_count"
  | "new_dislike_count"

const metricConfig: ReadonlyArray<{
  key: MetricKey
  name: string
  color: string
}> = [
  {
    key: "active_user_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.931f9c9d4123"
    ),
    color: "var(--color-chart-3)",
  },
  {
    key: "new_user_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.ebabc83b6830"
    ),
    color: "var(--color-chart-4)",
  },
  {
    key: "new_review_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.0b50583f022e"
    ),
    color: "var(--color-chart-5)",
  },
  {
    key: "new_point_amount",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.a7d035e5fecb"
    ),
    color: "var(--color-chart-2)",
  },
  {
    key: "review_author_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.c1320b7dd4f9"
    ),
    color: "var(--color-primary)",
  },
  {
    key: "new_like_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.e1a2345cdc47"
    ),
    color: "var(--color-chart-1)",
  },
  {
    key: "new_dislike_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.77d13c083dba"
    ),
    color: "var(--color-destructive)",
  },
  {
    key: "total_user_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.545a676083c8"
    ),
    color: "var(--color-chart-1)",
  },
  {
    key: "total_review_count",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.c5ddc469e1d4"
    ),
    color: "var(--color-chart-2)",
  },
  {
    key: "reviewed_course_total",
    name: copy(
      "copy.user.components.site-stats.daily-stats-chart.51f48f6d74a7"
    ),
    color: "var(--color-chart-5)",
  },
]

const chartHeight = 192

export function DailyStatsChart({ stats }: DailyStatsChartProps) {
  const chartData = [...stats].sort((a, b) =>
    a.stat_date.localeCompare(b.stat_date)
  )

  function formatStatDateLabel(label: unknown): string {
    return formatDateInputValue(typeof label === "string" ? label : "")
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {metricConfig.map((metric) => (
        <Card key={metric.key} size="sm" className="min-w-0">
          <CardHeader>
            <CardTitle>{metric.name}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-48 w-full min-w-0">
              <ResponsiveContainer
                width="100%"
                height="100%"
                minHeight={chartHeight}
                initialDimension={{ width: 0, height: chartHeight }}
              >
                <LineChart
                  data={chartData}
                  margin={{ top: 4, right: 12, left: 0, bottom: 4 }}
                >
                  <CartesianGrid
                    stroke="var(--color-border)"
                    strokeDasharray="3 3"
                    vertical={false}
                  />
                  <XAxis
                    dataKey="stat_date"
                    tickFormatter={formatDateInputValue}
                    tick={{
                      fontSize: 11,
                      fill: "var(--color-muted-foreground)",
                    }}
                    minTickGap={18}
                  />
                  <YAxis
                    tick={{
                      fontSize: 11,
                      fill: "var(--color-muted-foreground)",
                    }}
                    width={36}
                  />
                  <Tooltip
                    labelFormatter={formatStatDateLabel}
                    contentStyle={{
                      borderRadius: 12,
                      border: "1px solid var(--color-border)",
                      backgroundColor: "var(--color-card)",
                      color: "var(--color-card-foreground)",
                    }}
                  />
                  <Line
                    type="monotone"
                    dataKey={metric.key}
                    name={metric.name}
                    stroke={metric.color}
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
