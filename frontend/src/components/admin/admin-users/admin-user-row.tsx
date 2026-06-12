import type { AdminUserDTO } from "@/api/admin-user"
import { TableCell, TableRow } from "@/components/ui/table"
import { formatDateTime } from "@/lib/date"
import { RevokeAdminIconDialog } from "./revoke-admin-icon-dialog"

interface AdminUserRowProps {
  admin: AdminUserDTO
  currentUserID: number
  currentUserIsSuperAdmin: boolean
  isRevoking: boolean
  onRevoke: (userID: number) => void
}

export function AdminUserRow({
  admin,
  currentUserID,
  currentUserIsSuperAdmin,
  isRevoking,
  onRevoke,
}: AdminUserRowProps) {
  return (
    <TableRow>
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
          <RevokeAdminIconDialog
            username={admin.username}
            disabled={isRevoking}
            onConfirm={() => onRevoke(admin.id)}
          />
        )}
      </TableCell>
    </TableRow>
  )
}
