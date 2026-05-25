import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useAdminUsers, useRevokeAdminUser } from "@/hooks/use-admin-user"
import { formatDateTime } from "@/lib/date"

interface AdminUsersTabProps {
  currentUserID: number
}

export function AdminUsersTab({ currentUserID }: AdminUsersTabProps) {
  const adminsQuery = useAdminUsers()
  const revokeAdminMutation = useRevokeAdminUser()

  async function revokeAdmin(userID: number) {
    await revokeAdminMutation.mutateAsync(userID)
  }

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-lg font-medium">当前管理员</h2>
        {adminsQuery.data ? (
          <p className="text-sm text-muted-foreground">
            共 {adminsQuery.data.length} 人
          </p>
        ) : null}
      </div>
      {adminsQuery.isLoading ? (
        <Skeleton className="h-32 w-full" />
      ) : adminsQuery.data && adminsQuery.data.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>用户名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>活跃时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {adminsQuery.data.map((admin) => (
              <TableRow key={admin.id}>
                <TableCell className="font-mono">{admin.id}</TableCell>
                <TableCell className="font-mono">{admin.username}</TableCell>
                <TableCell>{admin.email || "-"}</TableCell>
                <TableCell>{formatDateTime(admin.last_seen_at)}</TableCell>
                <TableCell className="text-right">
                  {admin.id === currentUserID ? (
                    <span className="inline-flex h-8 items-center text-sm text-muted-foreground">
                      当前用户
                    </span>
                  ) : (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => revokeAdmin(admin.id)}
                      disabled={revokeAdminMutation.isPending}
                    >
                      撤销权限
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : (
        <p className="text-sm text-muted-foreground">暂无管理员</p>
      )}
    </section>
  )
}
