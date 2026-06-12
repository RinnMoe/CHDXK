import { Badge } from "@/components/ui/badge"
import type { AdminUserDTO } from "@/api/admin-user"
import { formatDateTime, formatNullableDateTime } from "@/lib/date"

function SuspensionText({
  suspendedAt,
  suspendTill,
}: {
  suspendedAt?: string
  suspendTill?: string
}) {
  if (!suspendedAt) return <span className="text-muted-foreground">未封禁</span>

  return (
    <span>
      {formatDateTime(suspendedAt)} 至{" "}
      {formatNullableDateTime(suspendTill, "未设置")}
    </span>
  )
}

export function AdminUserStatus({ user }: { user: AdminUserDTO }) {
  return (
    <div className="grid gap-x-8 gap-y-3 text-sm sm:grid-cols-2 lg:grid-cols-3">
      <div>
        <p className="text-muted-foreground">用户 ID</p>
        <p className="font-mono">{user.id}</p>
      </div>
      <div>
        <p className="text-muted-foreground">用户名</p>
        <p className="font-mono">{user.username}</p>
      </div>
      <div>
        <p className="text-muted-foreground">邮箱</p>
        <p>{user.email}</p>
      </div>
      <div>
        <p className="text-muted-foreground">角色</p>
        <p>{user.role}</p>
      </div>
      <div>
        <p className="text-muted-foreground">注册时间</p>
        <p>{formatDateTime(user.created_at)}</p>
      </div>
      <div>
        <p className="text-muted-foreground">活跃时间</p>
        <p>{formatDateTime(user.last_seen_at)}</p>
      </div>
      <div className="sm:col-span-2 lg:col-span-3">
        <p className="text-muted-foreground">密码</p>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          <Badge variant={user.password_hash ? "secondary" : "outline"}>
            {user.password_hash ? "已设置" : "未设置"}
          </Badge>
          {user.password_hash ? (
            <span className="max-w-full font-mono text-xs break-all text-muted-foreground">
              {user.password_hash}
            </span>
          ) : null}
        </div>
      </div>
      <div className="sm:col-span-2 lg:col-span-3">
        <p className="text-muted-foreground">封禁状态</p>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          <Badge variant={user.suspended ? "destructive" : "secondary"}>
            {user.suspended ? "已封禁" : "正常"}
          </Badge>
          <SuspensionText
            suspendedAt={user.suspended_at}
            suspendTill={user.suspend_till}
          />
        </div>
      </div>
    </div>
  )
}
