import { useEffect } from "react"
import { Link, useSearchParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import {
  COURSE_ENROLLMENT_SYNC_CHANNEL,
  type CourseEnrollmentSyncMessage,
} from "@/lib/course-enrollment-sync"

export function CourseEnrollmentSyncCallbackPage() {
  const [searchParams] = useSearchParams()
  const status = searchParams.get("status") === "ok" ? "ok" : "error"
  const semester = searchParams.get("semester") ?? ""
  const message = searchParams.get("message") ?? undefined
  const matched = Number(searchParams.get("matched") ?? "0") || 0
  const total = Number(searchParams.get("total") ?? "0") || 0

  useEffect(() => {
    const payload: CourseEnrollmentSyncMessage = {
      status,
      semester,
      message,
      matched,
      total,
    }
    const channel = new BroadcastChannel(COURSE_ENROLLMENT_SYNC_CHANNEL)
    channel.postMessage(payload)
    window.opener?.postMessage(
      { type: COURSE_ENROLLMENT_SYNC_CHANNEL, payload },
      window.location.origin
    )
    const timer = window.setTimeout(() => window.close(), 900)
    return () => {
      window.clearTimeout(timer)
      channel.close()
    }
  }, [matched, message, semester, status, total])

  return (
    <div className="flex min-h-dvh items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-4 text-center">
        <h1 className="text-xl font-semibold">
          {status === "ok" ? "同步完成" : "同步失败"}
        </h1>
        <p className="text-sm text-muted-foreground">
          {status === "ok"
            ? `已同步 ${semester || "所选学期"}，匹配 ${matched} 条记录。`
            : (message ?? "请关闭窗口后重试。")}
        </p>
        <Button asChild variant="outline">
          <Link
            to={`/course/mine?type=enrolled${semester ? `&semester=${semester}` : ""}`}
          >
            返回我的课程
          </Link>
        </Button>
      </div>
    </div>
  )
}
