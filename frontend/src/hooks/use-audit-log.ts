import { useQuery } from "@tanstack/react-query"
import { listAuditLogs, type AuditLogListFilter } from "@/api/audit-log"

export function useAuditLogs(filter: AuditLogListFilter) {
  return useQuery({
    queryKey: ["audit-logs", filter],
    queryFn: () => listAuditLogs(filter),
  })
}
