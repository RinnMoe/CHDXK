import { RiShieldCrossLine } from "@remixicon/react"
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
  currentUserIsSuperAdmin: boolean
}

export function AdminUsersTab({
  currentUserID,
  currentUserIsSuperAdmin,
}: AdminUsersTabProps) {
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
              <TableHead>角色</TableHead>
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
                <TableCell>{admin.role}</TableCell>
                <TableCell>{formatDateTime(admin.last_seen_at)}</TableCell>
                <TableCell className="text-right">
                  {admin.id === currentUserID ? (
                    <span className="inline-flex h-8 items-center text-sm text-muted-foreground">
                      当前用户
                    </span>
                  ) : admin.is_super_admin() ? (
                    <span className="inline-flex h-8 items-center text-sm text-muted-foreground">
                      超级管理员
                    </span>
                  ) : !currentUserIsSuperAdmin ? (
                    <span className="inline-flex h-8 items-center text-sm text-muted-foreground">
                      需要超级管理员权限
                    </span>
                  ) : (
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          className="hover:bg-destructive/10 hover:text-destructive focus-visible:border-destructive/40 focus-visible:ring-destructive/20 dark:hover:bg-destructive/20"
                          aria-label={`撤销 ${admin.username} 的管理员权限`}
                          disabled={revokeAdminMutation.isPending}
                        >
                          <RiShieldCrossLine />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent size="sm">
                        <AlertDialogHeader>
                          <AlertDialogTitle>撤销管理员权限</AlertDialogTitle>
                          <AlertDialogDescription>
                            撤销 {admin.username} 的管理员权限。
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>取消</AlertDialogCancel>
                          <AlertDialogAction
                            variant="destructive"
                            onClick={() => {
                              void revokeAdmin(admin.id)
                            }}
                          >
                            撤销
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
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
