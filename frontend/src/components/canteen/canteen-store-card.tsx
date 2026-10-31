import { Link } from "@tanstack/react-router"

import type { CanteenStoreListItemDTO } from "@/api/canteen"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"

export function CanteenStoreCard({
  store,
}: {
  store: CanteenStoreListItemDTO
}) {
  return (
    <Link
      to="/canteen/stores/$storeID"
      params={{ storeID: String(store.id) }}
      className="block rounded-xl focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none"
    >
      <Card size="sm" className="h-full transition-colors hover:bg-muted/40">
        <CardHeader>
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h3 className="truncate text-base font-semibold">{store.name}</h3>
              <p className="truncate text-xs text-muted-foreground">
                {store.hall_name}
                {store.floor ? ` · ${store.floor}` : ""}
              </p>
            </div>
            <div className="shrink-0 text-right">
              <div className="text-lg font-bold">
                {store.rating.score.toFixed(1)}
              </div>
              <div className="text-xs text-muted-foreground">
                {store.rating.count} 人评价
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {(store.confidence || store.summary_tags.length > 0) && (
            <div className="flex flex-wrap items-center gap-1.5">
              {store.confidence && (
                <Badge variant="outline">{store.confidence}</Badge>
              )}
              {store.summary_tags.slice(0, 3).map((tag) => (
                <Badge key={tag} variant="secondary">
                  {tag}
                </Badge>
              ))}
            </div>
          )}
          {store.categories.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {store.categories.map((category) => (
                <Badge key={category} variant="secondary">
                  {category}
                </Badge>
              ))}
            </div>
          )}
          <p className="text-xs text-muted-foreground">
            {store.review_count} 次到店记录
            {store.dish_count > 0
              ? ` · ${store.dish_count} 道菜品`
              : ""}
            {store.dish_photo_count > 0
              ? ` · ${store.dish_photo_count} 张实拍`
              : ""}
          </p>
        </CardContent>
      </Card>
    </Link>
  )
}
