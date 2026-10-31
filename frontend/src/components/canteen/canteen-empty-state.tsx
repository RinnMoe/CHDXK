import { Card, CardContent } from "@/components/ui/card"

export function CanteenEmptyState({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <Card size="sm" className="border-dashed bg-muted/20 shadow-none">
      <CardContent className="space-y-1 py-8 text-center">
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  )
}
