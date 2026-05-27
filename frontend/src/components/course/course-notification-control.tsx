import { Link } from "react-router-dom"
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
  0: "普通",
  1: "关注",
  2: "屏蔽",
}

const notificationDescriptions: Record<CourseNotificationLevel, string> = {
  0: "不关注，也不从点评列表中隐藏",
  1: "在关注课程和关注点评中查看",
  2: "在最新点评和点评搜索中隐藏此课点评",
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
        <span className="font-normal">通知等级：未登录</span>
        <Button asChild size="sm" variant="ghost" className="h-7 px-2">
          <Link to={`/login`}>登录</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
      <span className="font-normal">通知等级</span>
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
