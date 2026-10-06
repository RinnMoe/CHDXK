import { useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useAuth } from "@/contexts/auth-context"
import { getManagedUsers, manageUser } from "@/api/admin-users"
import type { AuthUserResponse } from "@/api/auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { UserManagementFields } from "./user-management-fields"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"

const roleName = (role: string) =>
  role === "super_admin" ? "超级管理员" : role === "admin" ? "管理员" : "用户"
type Operation =
  | "create"
  | "profile"
  | "password"
  | "role"
  | "moderation"
  | "points"
  | "delete"
type Editing = {
  action: Operation
  target?: AuthUserResponse
  requestID: string
}

export function UsersPanel() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [q, setQ] = useState("")
  const [role, setRole] = useState("")
  const [page, setPage] = useState(1)
  const [editing, setEditing] = useState<Editing | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState("")
  const [notice, setNotice] = useState("")
  const query = useQuery({
    queryKey: ["admin-users", q, role, page],
    queryFn: () => getManagedUsers(q, role, page),
  })
  const [localRole, setLocalRole] = useState("user")
  function open(action: Operation, target?: AuthUserResponse) {
    setEditing({ action, target, requestID: crypto.randomUUID() })
    setLocalRole(target?.local_role || "user")
    setError("")
    setNotice("")
  }
  async function submit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editing) return
    const form = new FormData(event.currentTarget)
    const get = (key: string) => (form.get(key) as string | null) ?? ""
    let body: unknown
    switch (editing.action) {
      case "create":
      case "profile":
        body = {
          username: get("username"),
          nickname: get("nickname"),
          password: get("password"),
        }
        break
      case "password":
        body = { password: get("password") }
        break
      case "role":
        body = { role: localRole, permissions: form.getAll("permissions") }
        break
      case "moderation":
        body = {
          action: get("action"),
          reason: get("reason"),
          until: get("until") ? new Date(get("until")).toISOString() : null,
        }
        break
      case "points":
        body = {
          amount:
            Number(get("amount")) * (get("direction") === "reclaim" ? -1 : 1),
          reason: get("reason"),
          request_id: editing.requestID,
        }
        break
    }
    setBusy(true)
    setError("")
    try {
      const result = await manageUser(editing.target?.id, editing.action, body)
      setNotice(
        result.balance === undefined
          ? "操作已完成"
          : `积分已调整，资料站余额为 ${result.balance}`
      )
      setEditing(null)
      await queryClient.invalidateQueries({ queryKey: ["admin-users"] })
    } catch (e) {
      setError(e instanceof Error ? e.message : "操作失败，请重试")
    } finally {
      setBusy(false)
    }
  }
  const titles: Record<Operation, string> = {
    create: "新增本站账号",
    profile: "修改用户资料",
    password: "重置本站密码",
    role: "设置身份与权限",
    moderation: "处理用户",
    points: "调整积分",
    delete: "删除本站账号",
  }
  const inputClass = "h-9 rounded-md border bg-background px-3 text-sm"
  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">用户管理</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            查询账号并处理本站违规行为。
          </p>
        </div>
        {user?.can("users") && (
          <Button onClick={() => open("create")}>新增账号</Button>
        )}
      </div>
      <div className="flex flex-wrap gap-3">
        <Input
          className="max-w-sm"
          aria-label="搜索账号或昵称"
          placeholder="搜索账号名或昵称"
          value={q}
          onChange={(e) => {
            setQ(e.target.value)
            setPage(1)
          }}
        />
        <select
          aria-label="用户身份"
          className={inputClass}
          value={role}
          onChange={(e) => {
            setRole(e.target.value)
            setPage(1)
          }}
        >
          <option value="">全部身份</option>
          <option value="super_admin">超级管理员</option>
          <option value="admin">管理员</option>
          <option value="user">用户</option>
        </select>
      </div>
      {notice && (
        <p role="status" className="text-sm">
          {notice}
        </p>
      )}
      {query.isPending && <p>加载中…</p>}
      {query.error && (
        <p role="alert" className="text-destructive">
          {query.error.message}
        </p>
      )}
      <div className="overflow-x-auto rounded-lg border">
        <table className="w-full text-left text-sm">
          <thead className="bg-muted/40">
            <tr>
              {["用户", "身份", "资料站", "状态", "操作"].map((label) => (
                <th key={label} className="p-3 font-medium whitespace-nowrap">
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {query.data?.items.map((target) => {
              const protectedUser =
                target.id === 1 ||
                target.id === user?.id ||
                !!target.permissions?.some((p) => !user?.can(p)) ||
                (target.role === "admin" &&
                  !user?.is_super_admin() &&
                  !user?.can("roles"))
              const banned =
                !!target.suspended_at &&
                (!target.suspend_till ||
                  new Date(target.suspend_till) > new Date())
              const muted =
                !!target.muted_until &&
                new Date(target.muted_until) > new Date()
              return (
                <tr key={target.id} className="border-t align-top">
                  <td className="p-3">
                    <div className="font-medium">
                      {target.nickname || target.username}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {target.username} · #{target.id}
                    </div>
                  </td>
                  <td className="p-3 whitespace-nowrap">
                    {roleName(target.role)}
                  </td>
                  <td className="p-3">
                    {target.subject
                      ? target.bound_nickname || "已绑定"
                      : "未绑定"}
                  </td>
                  <td className="p-3">
                    <div>{banned ? "已封禁" : muted ? "已禁言" : "正常"}</div>
                    {target.moderation_reason && (
                      <p className="mt-1 max-w-48 text-xs wrap-break-word text-muted-foreground">
                        {target.moderation_reason}
                      </p>
                    )}
                    {target.must_change_password && (
                      <p className="text-xs text-muted-foreground">
                        待修改初始密码
                      </p>
                    )}
                  </td>
                  <td className="p-2">
                    <div className="flex min-w-36 flex-wrap gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={protectedUser}
                        onClick={() => open("moderation", target)}
                      >
                        处理
                      </Button>
                      {user?.can("users") && (
                        <>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={protectedUser}
                            onClick={() => open("profile", target)}
                          >
                            编辑
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={protectedUser}
                            onClick={() => open("password", target)}
                          >
                            重置密码
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={protectedUser}
                            onClick={() => open("delete", target)}
                          >
                            删除
                          </Button>
                        </>
                      )}
                      {user?.can("roles") && (
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={protectedUser}
                          onClick={() => open("role", target)}
                        >
                          权限
                        </Button>
                      )}
                      {user?.can("points") && (
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={
                            !target.subject || !target.has_points_account
                          }
                          onClick={() => open("points", target)}
                        >
                          积分
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {query.data?.items.length === 0 && (
          <p className="p-8 text-center text-muted-foreground">
            没有符合条件的用户
          </p>
        )}
      </div>
      <div className="flex items-center justify-between gap-2 text-sm">
        <span>共 {query.data?.total || 0} 个账号</span>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={page === 1}
            onClick={() => setPage(page - 1)}
          >
            上一页
          </Button>
          <span>第 {page} 页</span>
          <Button
            size="sm"
            variant="outline"
            disabled={page * 20 >= (query.data?.total || 0)}
            onClick={() => setPage(page + 1)}
          >
            下一页
          </Button>
        </div>
      </div>
      <Dialog
        open={!!editing}
        onOpenChange={(value) => {
          if (!value && !busy) setEditing(null)
        }}
      >
        <DialogContent className="max-h-[90svh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editing && titles[editing.action]}</DialogTitle>
            <DialogDescription>
              {editing?.target
                ? `${editing.target.nickname || editing.target.username} · #${editing.target.id}`
                : "创建可使用本站账号名和密码登录的账号。"}
            </DialogDescription>
          </DialogHeader>
          {editing && (
            <form
              key={`${editing.action}:${editing.target?.id}`}
              onSubmit={submit}
              className="space-y-4"
            >
              <UserManagementFields
                editing={editing}
                user={user}
                localRole={localRole}
                setLocalRole={setLocalRole}
              />
              {error && (
                <p role="alert" className="text-sm text-destructive">
                  {error}
                </p>
              )}
              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  disabled={busy}
                  onClick={() => setEditing(null)}
                >
                  取消
                </Button>
                <Button
                  type="submit"
                  disabled={busy}
                  variant={
                    editing.action === "delete" ? "destructive" : "default"
                  }
                >
                  {busy ? "正在处理…" : "确认"}
                </Button>
              </div>
            </form>
          )}
        </DialogContent>
      </Dialog>
    </section>
  )
}
