import type { CanteenDimensionsDTO } from "@/api/canteen"

const dimensionRows = [
  { key: "taste", label: "味道" },
  { key: "value", label: "性价比" },
  { key: "queue", label: "排队效率" },
  { key: "hygiene", label: "卫生环境" },
  { key: "service", label: "服务态度" },
] as const

export function CanteenDimensionBars({
  dimensions,
}: {
  dimensions: CanteenDimensionsDTO
}) {
  return (
    <div className="space-y-2">
      {dimensionRows.map((row) => {
        const value = dimensions[row.key]
        return (
          <div key={row.key} className="flex items-center gap-3 text-sm">
            <span className="w-16 shrink-0 text-muted-foreground">
              {row.label}
            </span>
            <div className="h-2 flex-1 overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary"
                style={{ width: `${Math.max(0, Math.min(5, value)) * 20}%` }}
              />
            </div>
            <span className="w-8 text-right tabular-nums">
              {value > 0 ? value.toFixed(1) : "-"}
            </span>
          </div>
        )
      })}
    </div>
  )
}
