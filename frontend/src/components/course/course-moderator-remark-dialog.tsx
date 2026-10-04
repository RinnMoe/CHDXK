import { copy } from "@/lib/copy"
import { useState, type SyntheticEvent } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import type { CourseDetailDTO } from "@/api/course"
import { useUpdateCourseModeratorRemark } from "@/hooks/use-course"

interface CourseModeratorRemarkDialogProps {
  course: CourseDetailDTO
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CourseModeratorRemarkDialog({
  course,
  open,
  onOpenChange,
}: CourseModeratorRemarkDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <CourseModeratorRemarkForm
          key={`${course.id}-${open ? (course.moderator_remark ?? "") : "closed"}`}
          course={course}
          onOpenChange={onOpenChange}
        />
      </DialogContent>
    </Dialog>
  )
}

function CourseModeratorRemarkForm({
  course,
  onOpenChange,
}: Pick<CourseModeratorRemarkDialogProps, "course" | "onOpenChange">) {
  const [remark, setRemark] = useState(course.moderator_remark ?? "")
  const [error, setError] = useState<string | null>(null)
  const { mutateAsync, isPending } = useUpdateCourseModeratorRemark()

  async function handleSubmit(e: SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    try {
      await mutateAsync({
        courseID: course.id,
        cmd: { moderator_remark: remark },
      })
      onOpenChange(false)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : copy(
              "copy.user.components.course.course-moderator-remark-dialog.6309a3bb5ba4"
            )
      )
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <DialogHeader>
        <DialogTitle>
          {copy(
            "copy.user.components.course.course-moderator-remark-dialog.8a9f367b57e3"
          )}
        </DialogTitle>
        <DialogDescription>
          {course.code} {course.name}
          {copy(
            "copy.user.components.course.course-moderator-remark-dialog.9a67a794544d"
          )}
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-2">
        <Label htmlFor={`course-moderator-remark-${course.id}`}>
          {copy(
            "copy.user.components.course.course-moderator-remark-dialog.2e87a200b286"
          )}
        </Label>
        <Textarea
          id={`course-moderator-remark-${course.id}`}
          value={remark}
          onChange={(e) => setRemark(e.target.value)}
          rows={5}
          className="resize-y text-sm"
          placeholder={copy(
            "copy.user.components.course.course-moderator-remark-dialog.d34df2f6b18b"
          )}
        />
      </div>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      <DialogFooter>
        <Button
          type="button"
          variant="ghost"
          onClick={() => onOpenChange(false)}
        >
          {copy(
            "copy.user.components.course.course-moderator-remark-dialog.2cd0f3be8738"
          )}
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending
            ? copy(
                "copy.user.components.course.course-moderator-remark-dialog.917b1c1f18d0"
              )
            : copy(
                "copy.user.components.course.course-moderator-remark-dialog.4f8b0ab088ec"
              )}
        </Button>
      </DialogFooter>
    </form>
  )
}
