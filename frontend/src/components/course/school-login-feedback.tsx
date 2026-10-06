import { RiCheckboxCircleLine, RiLoader4Line } from "@remixicon/react"

const steps = ["扫码确认", "连接教务", "读取学期"]
const stages: Record<string, number> = {
  connecting: -1,
  waiting: -1,
  scanned: 0,
  authenticating: 1,
  discovering: 2,
  ready: 3,
}

export function SchoolLoginFeedback({ state }: { state: string }) {
  const current = stages[state]
  if (current === undefined) return null
  return (
    <div className="space-y-3">
      <ol
        aria-label="教务连接进度"
        className="flex flex-wrap gap-x-4 gap-y-2 text-sm"
      >
        {steps.map((label, index) => {
          const complete = index < current
          const active = index === current
          return (
            <li
              key={label}
              className={`flex items-center gap-1.5 ${complete ? "text-primary" : active ? "font-medium" : "text-muted-foreground"}`}
              aria-current={active ? "step" : undefined}
            >
              {complete ? (
                <RiCheckboxCircleLine className="size-4" aria-label="已完成" />
              ) : active ? (
                <RiLoader4Line
                  className="size-4 animate-spin motion-reduce:animate-none"
                  aria-label="进行中"
                />
              ) : (
                <span className="flex size-4 items-center justify-center rounded-full border text-xs">
                  {index + 1}
                </span>
              )}
              {label}
            </li>
          )
        })}
      </ol>
      {(state === "authenticating" || state === "discovering") && (
        <p className="text-sm text-muted-foreground">
          无需重复扫码，请等待页面自动更新。
        </p>
      )}
    </div>
  )
}
