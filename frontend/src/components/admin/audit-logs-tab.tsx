import { useSearchParams } from "react-router-dom"
import { RiSearchLine } from "@remixicon/react"
import { PaginationComponent } from "@/components/common/pagination"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useAuditLogs } from "@/hooks/use-audit-log"
import { formatDateTime } from "@/lib/date"
import { getErrorMessage, type FormSubmitEvent } from "./admin-utils"

const pageSize = 20
const allActionsValue = "__all__"

const actionOptions = [
  { value: "user.suspend", label: "封禁用户" },
  { value: "user.unsuspend", label: "解封用户" },
  { value: "admin.grant", label: "授予管理员" },
  { value: "admin.revoke", label: "撤回管理员" },
  { value: "system_api_key.create", label: "新增系统 API Key" },
  { value: "system_api_key.delete", label: "删除系统 API Key" },
  { value: "review.update", label: "修改点评" },
  { value: "review.delete", label: "删除点评" },
  { value: "review.moderator_remark.update", label: "修改管理备注" },
]

const actionLabels = Object.fromEntries(
  actionOptions.map((item) => [item.value, item.label])
)

function detailText(details: Record<string, unknown>) {
  const parts = Object.entries(details)
    .filter(([, value]) => value !== undefined && value !== null && value !== "")
    .map(([key, value]) => `${key}: ${String(value)}`)
  return parts.length ? parts.join("; ") : "-"
}

export function AuditLogsTab() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Math.max(1, Number(searchParams.get("audit_page") ?? "1") || 1)
  const startTime = searchParams.get("audit_start_time") ?? ""
  const endTime = searchParams.get("audit_end_time") ?? ""
  const action = searchParams.get("audit_action") ?? ""
  const actorUserID = Math.max(
    0,
    Number(searchParams.get("audit_actor_user_id") ?? "0") || 0
  )

  const logsQuery = useAuditLogs({
    start_time: startTime || undefined,
    end_time: endTime || undefined,
    action: action || undefined,
    actor_user_id: actorUserID || undefined,
    page,
    page_size: pageSize,
  })

  function handleSearch(event: FormSubmitEvent) {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const nextStart = String(formData.get("start_time") ?? "")
    const nextEnd = String(formData.get("end_time") ?? "")
    const nextActor = String(formData.get("actor_user_id") ?? "").trim()
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set("tab", "audit-log")
        next.set("audit_page", "1")
        setOrDelete(next, "audit_start_time", nextStart)
        setOrDelete(next, "audit_end_time", nextEnd)
        setOrDelete(next, "audit_actor_user_id", nextActor)
        return next
      },
      { replace: true }
    )
  }

  function setAction(value: string) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set("tab", "audit-log")
        next.set("audit_page", "1")
        if (value === allActionsValue) next.delete("audit_action")
        else next.set("audit_action", value)
        return next
      },
      { replace: true }
    )
  }

  function setPage(nextPage: number) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set("tab", "audit-log")
        next.set("audit_page", String(nextPage))
        return next
      },
      { replace: true }
    )
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <h2 className="text-lg font-medium">审计日志</h2>
      </div>

      <form className="flex flex-wrap items-end gap-2" onSubmit={handleSearch}>
        <div className="flex w-40 flex-col gap-1">
          <Label htmlFor="audit-start-time">开始时间</Label>
          <Input
            id="audit-start-time"
            name="start_time"
            type="date"
            defaultValue={startTime}
            className="w-full"
          />
        </div>
        <div className="flex w-40 flex-col gap-1">
          <Label htmlFor="audit-end-time">结束时间</Label>
          <Input
            id="audit-end-time"
            name="end_time"
            type="date"
            defaultValue={endTime}
            className="w-full"
          />
        </div>
        <div className="flex w-36 flex-col gap-1">
          <Label htmlFor="audit-actor-user-id">触发人 ID</Label>
          <Input
            id="audit-actor-user-id"
            name="actor_user_id"
            type="number"
            min={1}
            defaultValue={actorUserID || ""}
            className="w-full"
          />
        </div>
        <div className="flex w-52 flex-col gap-1">
          <Label htmlFor="audit-action">行为</Label>
          <Select value={action || allActionsValue} onValueChange={setAction}>
            <SelectTrigger id="audit-action" className="h-9 w-full items-center">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={allActionsValue}>全部行为</SelectItem>
              {actionOptions.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button type="submit" className="h-9">
          <RiSearchLine />
          查询
        </Button>
      </form>

      {logsQuery.data ? (
        <p className="text-sm text-muted-foreground">
          共 {logsQuery.data.total} 条
        </p>
      ) : null}

      {logsQuery.isLoading ? <Skeleton className="h-96 w-full" /> : null}

      {logsQuery.isError ? (
        <p className="text-sm text-destructive">
          {getErrorMessage(logsQuery.error)}
        </p>
      ) : null}

      {logsQuery.data && !logsQuery.isLoading ? (
        <div className="border-y">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>时间</TableHead>
                <TableHead>行为</TableHead>
                <TableHead>触发人</TableHead>
                <TableHead>对象</TableHead>
                <TableHead>详情</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logsQuery.data.items.length ? (
                logsQuery.data.items.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="whitespace-nowrap">
                      {formatDateTime(log.occurred_at)}
                    </TableCell>
                    <TableCell>{actionLabels[log.action] ?? log.action}</TableCell>
                    <TableCell className="font-mono">{log.actor_user_id}</TableCell>
                    <TableCell className="font-mono">
                      {log.target_type}:{log.target_id}
                    </TableCell>
                    <TableCell className="max-w-md text-muted-foreground">
                      {detailText(log.details)}
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={5} className="py-10 text-center text-muted-foreground">
                    暂无审计日志
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      ) : null}

      {logsQuery.data && logsQuery.data.total > pageSize ? (
        <PaginationComponent
          page={logsQuery.data.page}
          pageSize={logsQuery.data.page_size}
          total={logsQuery.data.total}
          onPageChange={setPage}
        />
      ) : null}
    </section>
  )
}

function setOrDelete(params: URLSearchParams, key: string, value: string) {
  if (value) params.set(key, value)
  else params.delete(key)
}
