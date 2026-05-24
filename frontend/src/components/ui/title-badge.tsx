import { cn } from "@/lib/utils"

interface TitleBadgeProps {
  children: string
  className?: string
}

export function TitleBadge({ children, className }: TitleBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex h-6 shrink-0 items-center rounded px-2 font-sans text-sm font-normal text-muted-foreground",
        className
      )}
    >
      {children}
    </span>
  )
}
