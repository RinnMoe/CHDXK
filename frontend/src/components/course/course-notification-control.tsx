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
  0: "不特别关注，也不屏蔽动态",
  1: "关注后可在关注动态中查看更新",
  2: "屏蔽后最新点评会过滤这门课",
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
        <span>通知等级：未登录</span>
        <Button asChild size="sm" variant="ghost" className="h-7 px-2">
          <Link to={`/login`}>登录</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
      <span>通知等级</span>
      <Select
        value={String(level)}
        onValueChange={handleChange}
        disabled={mutation.isPending}
      >
        <SelectTrigger size="sm" className="h-7 w-[112px]">
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
                <span>{notificationLabels[optionLevel]}</span>
                <span className="text-sm leading-4 text-muted-foreground">
                  {notificationDescriptions[optionLevel]}
                </span>
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button asChild size="sm" variant="ghost" className="h-7 px-2">
        <Link to="/courses/mine">查看全部</Link>
      </Button>
    </div>
  )
}
