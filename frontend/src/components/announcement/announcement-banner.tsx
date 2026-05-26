import { useState } from "react"
import { RiCloseLine, RiInformationLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/auth-context"
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
  const { user, isLoading } = useAuth()
  const { data } = useAnnouncements(Boolean(user))
  const [dismissed, setDismissed] = useState(() => getDismissed())

  if (isLoading || !user) return null

  const visible = (data ?? []).filter((a) => !dismissed.has(a.id))
  if (visible.length === 0) return null

  function dismiss(id: number) {
    const next = new Set(dismissed)
    next.add(id)
    setDismissed(next)
    persistDismissed(next)
  }

  return (
    <div className="mb-6 space-y-2">
      {visible.map((a) => {
        const title = a.title.trim()
        const body = a.body.trim()

        if (!title && !body) return null

        return (
          <div
            key={a.id}
            className="flex items-center gap-3 rounded-md border border-primary/20 bg-primary/5 px-4 py-3 text-sm"
          >
            <RiInformationLine className="size-4 shrink-0 text-muted-foreground" />
            <div className="min-w-0 flex-1">
              {title && <p className="font-medium text-primary">{title}</p>}
              {body && (
                <p
                  className={`text-sm whitespace-pre-line ${title ? "mt-1" : ""}`}
                >
                  {body}
                </p>
              )}
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-7 shrink-0"
              onClick={() => dismiss(a.id)}
              aria-label="dismiss"
            >
              <RiCloseLine className="size-4" />
            </Button>
          </div>
        )
      })}
    </div>
  )
}
