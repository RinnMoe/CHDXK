import { cn } from "@/lib/utils"

interface TitleBadgeProps {
  children: string
  className?: string
}

export function TitleBadge({ children, className }: TitleBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-5 shrink-0 items-center rounded px-1.5 text-xs font-medium text-muted-foreground",
        className,
      )}
    >
      {children}
    </span>
  )
}
