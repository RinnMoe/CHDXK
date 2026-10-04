import { copy } from "@/lib/copy"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
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
import { DateFilterPicker } from "./audit-logs/date-filter-picker"

function formDataString(value: FormDataEntryValue | null): string {
  return typeof value === "string" ? value : ""
}

const pageSize = 20
const allActionsValue = "__all__"
const routeApi = getRouteApi("/app/admin/user")

const actionOptions = [
  {
    value: "system_api_key.create",
    label: copy("copy.admin.components.admin.audit-logs-tab.4a36b70108bf"),
  },
  {
    value: "system_api_key.delete",
    label: copy("copy.admin.components.admin.audit-logs-tab.86fbfb0b1127"),
  },
  {
    value: "system_settings.update",
    label: copy("copy.admin.components.admin.audit-logs-tab.0cdb484341a7"),
  },
  {
    value: "review.update",
    label: copy("copy.admin.components.admin.audit-logs-tab.3cd5d9b5704c"),
  },
  {
    value: "review.delete",
    label: copy("copy.admin.components.admin.audit-logs-tab.e6b1e035bda1"),
  },
  {
    value: "review.moderator_remark.update",
    label: copy("copy.admin.components.admin.audit-logs-tab.c6b856679bb5"),
  },
]

const actionLabels = Object.fromEntries(
  actionOptions.map((item) => [item.value, item.label])
)

function detailText(details: Record<string, unknown>) {
  const parts = Object.entries(details)
    .filter(
      ([, value]) => value !== undefined && value !== null && value !== ""
    )
    .map(([key, value]) => `${key}: ${String(value)}`)
  return parts.length ? parts.join("; ") : "-"
}

export function AuditLogsTab() {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/admin/user" })
  const page = Math.max(1, search.audit_page ?? 1)
  const startTime = search.audit_start_time ?? ""
  const endTime = search.audit_end_time ?? ""
  const action = search.audit_action ?? ""
  const actorUserID = Math.max(0, search.audit_actor_user_id ?? 0)

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
    const nextStart = formDataString(formData.get("start_time"))
    const nextEnd = formDataString(formData.get("end_time"))
    const nextActor = formDataString(formData.get("actor_user_id")).trim()
    const nextActorID = Number(nextActor)
    const nextActorUserID =
      nextActor && Number.isFinite(nextActorID) && nextActorID > 0
        ? Math.trunc(nextActorID)
        : undefined
    void navigate({
      search: (prev) => ({
        ...prev,
        tab: "audit-log",
        audit_page: 1,
        audit_start_time: nextStart || undefined,
        audit_end_time: nextEnd || undefined,
        audit_actor_user_id: nextActorUserID,
      }),
      replace: true,
      resetScroll: false,
    })
  }

  function setAction(value: string) {
    void navigate({
      search: (prev) => ({
        ...prev,
        tab: "audit-log",
        audit_page: 1,
        audit_action: value === allActionsValue ? undefined : value,
      }),
      replace: true,
      resetScroll: false,
    })
  }

  function setDateFilter(key: "audit_start_time" | "audit_end_time") {
    return (value: string) => {
      void navigate({
        search: (prev) => ({
          ...prev,
          tab: "audit-log",
          audit_page: 1,
          [key]: value || undefined,
        }),
        replace: true,
        resetScroll: false,
      })
    }
  }

  function setPage(nextPage: number) {
    void navigate({
      search: (prev) => ({
        ...prev,
        tab: "audit-log",
        audit_page: nextPage,
      }),
      replace: true,
      resetScroll: false,
    })
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <h2 className="text-lg font-medium">
          {copy("copy.admin.components.admin.audit-logs-tab.a0f79e91f13f")}
        </h2>
      </div>

      <form className="flex flex-wrap items-end gap-2" onSubmit={handleSearch}>
        <DateFilterPicker
          id="audit-start-time"
          name="start_time"
          label={copy(
            "copy.admin.components.admin.audit-logs-tab.6a9906c79f26"
          )}
          value={startTime}
          onChange={setDateFilter("audit_start_time")}
        />
        <DateFilterPicker
          id="audit-end-time"
          name="end_time"
          label={copy(
            "copy.admin.components.admin.audit-logs-tab.f50276449943"
          )}
          value={endTime}
          onChange={setDateFilter("audit_end_time")}
        />
        <div className="flex w-36 flex-col gap-1">
          <Label htmlFor="audit-actor-user-id">
            {copy("copy.admin.components.admin.audit-logs-tab.8656460e0d46")}
          </Label>
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
          <Label htmlFor="audit-action">
            {copy("copy.admin.components.admin.audit-logs-tab.d92ffbeedfac")}
          </Label>
          <Select value={action || allActionsValue} onValueChange={setAction}>
            <SelectTrigger
              id="audit-action"
              className="h-9 w-full items-center"
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={allActionsValue}>
                {copy(
                  "copy.admin.components.admin.audit-logs-tab.ae90ef46a90c"
                )}
              </SelectItem>
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
          {copy("copy.admin.components.admin.audit-logs-tab.bcd6771e08ec")}
        </Button>
      </form>

      {logsQuery.data ? (
        <p className="text-sm text-muted-foreground">
          {copy("copy.admin.components.admin.audit-logs-tab.2428420e407f")}
          {logsQuery.data.total}
          {copy("copy.admin.components.admin.audit-logs-tab.6fb2840d568a")}
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
                <TableHead>
                  {copy(
                    "copy.admin.components.admin.audit-logs-tab.8b6ff498515b"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.admin.components.admin.audit-logs-tab.d92ffbeedfac"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.admin.components.admin.audit-logs-tab.814362ba214f"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.admin.components.admin.audit-logs-tab.53f92c063963"
                  )}
                </TableHead>
                <TableHead>
                  {copy(
                    "copy.admin.components.admin.audit-logs-tab.979a332955c8"
                  )}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logsQuery.data.items.length ? (
                logsQuery.data.items.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="whitespace-nowrap">
                      {formatDateTime(log.occurred_at)}
                    </TableCell>
                    <TableCell>
                      {actionLabels[log.action] ?? log.action}
                    </TableCell>
                    <TableCell className="font-mono">
                      {log.actor_user_id}
                    </TableCell>
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
                  <TableCell
                    colSpan={5}
                    className="py-10 text-center text-muted-foreground"
                  >
                    {copy(
                      "copy.admin.components.admin.audit-logs-tab.ea97b4577750"
                    )}
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
