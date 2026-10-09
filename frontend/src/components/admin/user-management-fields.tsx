import type { AuthUserDTO, AuthUserResponse } from "@/api/auth"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const capabilities = [
  { key: "users", label: "用户增删改与密码重置" },
  { key: "roles", label: "用户身份与权限分配" },
  { key: "points", label: "资料站积分发放与收回" },
  { key: "settings", label: "站点配置" },
  { key: "course_data", label: "课程数据（教师资料、开课记录与教务同步）" },
  { key: "maintenance", label: "开发维护（接口密钥与操作记录）" },
]
export function UserManagementFields({
  editing,
  user,
  localRole,
  setLocalRole,
}: {
  editing: { action: string; target?: AuthUserResponse }
  user: AuthUserDTO | null | undefined
  localRole: string
  setLocalRole: (role: string) => void
}) {
  const inputClass = "h-9 rounded-md border bg-background px-3 text-sm"
  const inheritedAdmin = [
    "admin",
    "administrator",
    "superadmin",
    "super_admin",
    "super-admin",
  ].includes(editing.target?.identity_role?.toLowerCase() ?? "")
  const canReceivePermissions = localRole === "admin" || inheritedAdmin
  return (
    <>
      {(editing.action === "create" || editing.action === "profile") && (
        <>
          <div className="space-y-2">
            <Label htmlFor="managed-username">账号名</Label>
            <Input
              id="managed-username"
              name="username"
              defaultValue={editing.target?.username}
              required
              minLength={3}
              maxLength={32}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="managed-nickname">昵称</Label>
            <Input
              id="managed-nickname"
              name="nickname"
              defaultValue={editing.target?.nickname}
              required
              maxLength={64}
            />
          </div>
        </>
      )}
      {(editing.action === "create" || editing.action === "password") && (
        <>
          <div className="space-y-2">
            <Label htmlFor="managed-password">初始密码</Label>
            <Input
              id="managed-password"
              name="password"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
            />
          </div>
          <p className="text-sm text-muted-foreground">
            至少 8 个字符。用户下次登录必须修改密码；原登录会话将失效。
          </p>
        </>
      )}
      {editing.action === "role" && (
        <>
          <div className="space-y-2">
            <Label htmlFor="managed-role">本站身份</Label>
            <select
              id="managed-role"
              className={`${inputClass} w-full`}
              value={localRole}
              onChange={(e) => setLocalRole(e.target.value)}
            >
              <option value="user">用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          {inheritedAdmin && (
            <p className="text-sm text-muted-foreground">
              此账号从资料站继承管理员身份。取消本站管理员身份后，仍具有继承的管理员身份。
            </p>
          )}
          <fieldset
            disabled={!canReceivePermissions}
            className="space-y-3 rounded-lg border p-3 disabled:opacity-60"
          >
            <legend className="px-1 text-sm font-medium">额外管理权限</legend>
            <p className="text-xs text-muted-foreground">
              {canReceivePermissions
                ? "管理员默认可维护内容及处理违规用户。勾选下列功能后，可执行对应管理操作。"
                : "请先将本站身份设为管理员，再授予额外管理权限。"}
            </p>
            {capabilities
              .filter((p) => user?.can(p.key))
              .map((p) => (
                <label key={p.key} className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    name="permissions"
                    value={p.key}
                    defaultChecked={editing.target?.permissions?.includes(
                      p.key
                    )}
                  />
                  {p.label}
                </label>
              ))}
          </fieldset>
        </>
      )}
      {editing.action === "moderation" && (
        <>
          <div className="space-y-2">
            <Label htmlFor="managed-action">处理方式</Label>
            <select
              id="managed-action"
              name="action"
              className={`${inputClass} w-full`}
            >
              <option value="mute">禁言</option>
              <option value="unmute">解除禁言</option>
              <option value="ban">封禁</option>
              <option value="unban">解除封禁</option>
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="managed-until">结束时间（可选）</Label>
            <Input id="managed-until" type="datetime-local" name="until" />
            <p className="text-xs text-muted-foreground">
              留空表示长期处理。仅影响本站，不影响资料站账号。
            </p>
          </div>
        </>
      )}
      {editing.action === "points" && (
        <>
          <p className="text-sm text-muted-foreground">
            调整该账号绑定的资料站积分余额。绑定同一资料站账号的本站账号共用此余额。
          </p>
          <div className="flex gap-2">
            <select
              aria-label="积分操作"
              name="direction"
              className={inputClass}
            >
              <option value="grant">发放</option>
              <option value="reclaim">收回</option>
            </select>
            <Input
              aria-label="积分数量"
              name="amount"
              type="number"
              min={1}
              max={1000000}
              step={1}
              required
            />
          </div>
        </>
      )}
      {(editing.action === "moderation" || editing.action === "points") && (
        <div className="space-y-2">
          <Label htmlFor="managed-reason">原因</Label>
          <Input
            id="managed-reason"
            name="reason"
            required
            maxLength={editing.action === "points" ? 200 : 500}
          />
        </div>
      )}
      {editing.action === "delete" && (
        <p className="text-sm">
          删除后，该账号将无法登录；已发布点评保留。资料站账号不会被删除。
        </p>
      )}
    </>
  )
}
