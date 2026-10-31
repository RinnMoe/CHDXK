import { canteenTags, maxCanteenTags } from "@/lib/canteen-tags"
import { Badge } from "@/components/ui/badge"

export function CanteenTagPicker({
  selected,
  onChange,
}: {
  selected: string[]
  onChange: (tags: string[]) => void
}) {
  function toggle(tag: string) {
    if (selected.includes(tag)) {
      onChange(selected.filter((item) => item !== tag))
      return
    }
    if (selected.length >= maxCanteenTags) return
    onChange([...selected, tag])
  }

  return (
    <div className="flex flex-wrap gap-2">
      {canteenTags.map((tag) => {
        const active = selected.includes(tag)
        return (
          <button
            key={tag}
            type="button"
            onClick={() => toggle(tag)}
            aria-pressed={active}
            disabled={!active && selected.length >= maxCanteenTags}
            className="rounded-full focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none disabled:opacity-50"
          >
            <Badge variant={active ? "default" : "outline"}>{tag}</Badge>
          </button>
        )
      })}
    </div>
  )
}
