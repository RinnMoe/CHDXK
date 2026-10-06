import type { ReactNode } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"

interface ReviewDialogProps {
  title: string
  course?: { name: string; main_teacher: { name: string } }
  onClose: () => void
  busy: boolean
  children: ReactNode
}

export function ReviewDialog({
  title,
  course,
  onClose,
  busy,
  children,
}: ReviewDialogProps) {
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy) onClose()
      }}
    >
      <DialogContent
        className="flex max-h-[90dvh] flex-col gap-4 overflow-hidden p-4 sm:max-w-3xl sm:p-6"
        showCloseButton={!busy}
        onInteractOutside={(event) => event.preventDefault()}
        onEscapeKeyDown={(event) => {
          if (busy) event.preventDefault()
        }}
      >
        <DialogHeader className="shrink-0 pr-8">
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription className="wrap-break-word">
            {course
              ? `${course.name} · ${course.main_teacher.name}`
              : "分享你的上课体验"}
          </DialogDescription>
        </DialogHeader>
        <div className="min-h-0 space-y-4 overflow-y-auto overscroll-contain pr-1">
          {children}
        </div>
      </DialogContent>
    </Dialog>
  )
}
