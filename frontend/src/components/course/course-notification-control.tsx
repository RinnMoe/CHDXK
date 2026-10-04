import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import {
  RiEyeCloseLine,
  RiNotification3Line,
  RiStarLine,
} from "@remixicon/react"
import type { CourseNotificationLevel } from "@/api/course"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useAuth } from "@/contexts/auth-context"
import { useSetNotificationLevel } from "@/hooks/use-course"

const notificationLabels: Record<CourseNotificationLevel, string> = {
  0: copy(
    "copy.user.components.course.course-notification-control.de907d10df98"
  ),
  1: copy(
    "copy.user.components.course.course-notification-control.eb137dee4c0a"
  ),
  2: copy(
    "copy.user.components.course.course-notification-control.5e2a947290d1"
  ),
}

const notificationDescriptions: Record<CourseNotificationLevel, string> = {
  0: copy(
    "copy.user.components.course.course-notification-control.b350488d45f0"
  ),
  1: copy(
    "copy.user.components.course.course-notification-control.a5e06fe853d2"
  ),
  2: copy(
    "copy.user.components.course.course-notification-control.fafd189cc3d0"
  ),
}

const notificationOptions = [
  { level: 0, Icon: RiNotification3Line },
  { level: 1, Icon: RiStarLine },
  { level: 2, Icon: RiEyeCloseLine },
] satisfies Array<{
  level: CourseNotificationLevel
  Icon: typeof RiNotification3Line
}>

interface CourseNotificationControlProps {
  courseID: number
  level: CourseNotificationLevel
}

export function CourseNotificationControl({
  courseID,
  level,
}: CourseNotificationControlProps) {
  const { user } = useAuth()
  const mutation = useSetNotificationLevel()

  function handleChange(value: string) {
    const nextLevel = Number(value) as CourseNotificationLevel
    mutation.mutate({ courseID, level: nextLevel })
  }

  if (!user) {
    return (
      <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
        <span className="font-normal">
          {copy(
            "copy.user.components.course.course-notification-control.1183e332583e"
          )}
        </span>
        <Button asChild size="sm" variant="ghost" className="h-7 px-2">
          <Link to="/login">
            {copy(
              "copy.user.components.course.course-notification-control.1e2df9c3075a"
            )}
          </Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
      <span className="font-normal">
        {copy(
          "copy.user.components.course.course-notification-control.4b5aba5d20da"
        )}
      </span>
      <Select
        value={String(level)}
        onValueChange={handleChange}
        disabled={mutation.isPending}
      >
        <SelectTrigger size="sm" className="h-7">
          <SelectValue>{notificationLabels[level]}</SelectValue>
        </SelectTrigger>
        <SelectContent className="min-w-64">
          {notificationOptions.map(({ level: optionLevel, Icon }) => (
            <SelectItem
              key={optionLevel}
              value={String(optionLevel)}
              textValue={notificationLabels[optionLevel]}
              className="items-start py-2"
            >
              <Icon className="mt-0.5" data-icon="inline-start" />
              <span className="flex min-w-0 flex-col gap-0.5">
                <span className="font-medium">
                  {notificationLabels[optionLevel]}
                </span>
                <span className="text-sm leading-4 text-muted-foreground">
                  {notificationDescriptions[optionLevel]}
                </span>
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
