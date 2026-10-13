import type { CourseListItemDTO } from "@/api/course"
import type { FilterItem } from "@/api/types"

export const discoveryBatchSize = 6

export function discoveryBatch(items: CourseListItemDTO[], batch: number) {
  const count = Math.ceil(items.length / discoveryBatchSize)
  const index = count > 0 ? batch % count : 0
  return items.slice(
    index * discoveryBatchSize,
    (index + 1) * discoveryBatchSize
  )
}

export function homeCategories(items: FilterItem[] = []) {
  return items
    .filter((item) => item.name.trim() && item.count > 0)
    .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
    .slice(0, 6)
}
