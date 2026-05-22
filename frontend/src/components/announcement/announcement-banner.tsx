import { useState } from "react"
import { RiCloseLine, RiInformationLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { useAnnouncements } from "@/hooks/use-announcement"

const STORAGE_KEY = "jcourse:dismissed-announcements"

function getDismissed(): Set<number> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return new Set()
    return new Set(JSON.parse(raw) as number[])
  } catch {
    return new Set()
  }
}

function persistDismissed(ids: Set<number>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(Array.from(ids)))
  } catch {
    // ignore
  }
}

export function AnnouncementBanner() {
  const { data } = useAnnouncements()
  const [dismissed, setDismissed] = useState(() => getDismissed())

  const visible = (data ?? []).filter((a) => !dismissed.has(a.id))
  if (visible.length === 0) return null

  function dismiss(id: number) {
    const next = new Set(dismissed)
    next.add(id)
    setDismissed(next)
    persistDismissed(next)
  }

  return (
    <div className="space-y-2 mt-4">
      {visible.map((a) => (
        <div
          key={a.id}
          className="flex items-start gap-3 rounded-md border bg-muted/50 px-4 py-3 text-sm"
        >
          <RiInformationLine className="size-4 mt-0.5 text-muted-foreground shrink-0" />
          <div className="min-w-0 flex-1">
            <p className="font-medium">{a.title}</p>
            <p className="text-muted-foreground mt-1 whitespace-pre-line">{a.body}</p>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-7 -mr-2 -mt-1 shrink-0"
            onClick={() => dismiss(a.id)}
            aria-label="dismiss"
          >
            <RiCloseLine className="size-4" />
          </Button>
        </div>
      ))}
    </div>
  )
}
