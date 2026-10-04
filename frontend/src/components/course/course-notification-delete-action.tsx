import { copy } from "@/lib/copy"
import { RiDeleteBinLine } from "@remixicon/react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import type { CourseListItemDTO } from "@/api/course"

export function CourseNotificationDeleteAction({
  course,
  recordName,
  onDelete,
}: {
  course: CourseListItemDTO
  recordName: string
  onDelete: () => void
}) {
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon-sm"
          className="hover:bg-destructive/10 hover:text-destructive focus-visible:border-destructive/40 focus-visible:ring-destructive/20 dark:hover:bg-destructive/20"
          aria-label={copy(
            "copy.user.components.course.course-notification-delete-action.bc5e6bd4be1c",
            { v0: recordName }
          )}
        >
          <RiDeleteBinLine />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent size="sm">
        <AlertDialogHeader>
          <AlertDialogTitle>
            {copy(
              "copy.user.components.course.course-notification-delete-action.2f9daa828907"
            )}
            {recordName}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {copy(
              "copy.user.components.course.course-notification-delete-action.b5e1bb61b32e"
            )}
            {course.name}
            {copy(
              "copy.user.components.course.course-notification-delete-action.7b60a692c160"
            )}
            {recordName}。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>
            {copy(
              "copy.user.components.course.course-notification-delete-action.2cd0f3be8738"
            )}
          </AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={onDelete}>
            {copy(
              "copy.user.components.course.course-notification-delete-action.2f9daa828907"
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
