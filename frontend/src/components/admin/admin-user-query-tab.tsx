import { useState } from "react"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import {
  RiKey2Line,
  RiLockLine,
  RiLockUnlockLine,
  RiSearchLine,
  RiShieldCrossLine,
  RiShieldUserLine,
} from "@remixicon/react"
import { EmailPrefixInput } from "@/components/auth/email-prefix-input"
import { PaginationComponent } from "@/components/common/pagination"
import { PointRecordList } from "@/components/point/point-record-list"
import { ReviewList } from "@/components/review/review-list"
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
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import {
  useAdminUserByEmail,
  useClearAdminUserSuspension,
  useGrantAdminUser,
  useResetAdminUserPassword,
  useRevokeAdminUser,
  useSuspendAdminUser,
} from "@/hooks/use-admin-user"
import { useUserPoints } from "@/hooks/use-point"
import { useUserReviews } from "@/hooks/use-review"
import { buildAuthEmail, normalizeAuthEmailPrefix } from "@/config/auth"
import { formatDateTime, formatNullableDateTime } from "@/lib/date"
import { getErrorMessage, type FormSubmitEvent } from "./admin-utils"

const reviewPageSize = 20
const pointPageSize = 20
const routeApi = getRouteApi("/app/admin/user")

function defaultPasswordFromEmail(email: string, username: string) {
  const source = email.split("@")[0] || username
  return source.trim().toLowerCase()
}

interface AdminUserQueryTabProps {
  currentUserID: number
  currentUserIsSuperAdmin: boolean
}

interface AdminUserEmailSearchFormProps {
  email: string
  onSearch: (email: string) => void
}

function AdminUserEmailSearchForm({
  email,
  onSearch,
}: AdminUserEmailSearchFormProps) {
  const [emailPrefix, setEmailPrefix] = useState(() =>
    normalizeAuthEmailPrefix(email)
  )

  function handleSubmit(event: FormSubmitEvent) {
    event.preventDefault()
    const nextPrefix = emailPrefix.trim()
    onSearch(nextPrefix ? buildAuthEmail(nextPrefix).toLowerCase() : "")
  }

  return (
    <form
      className="flex max-w-md flex-wrap items-end gap-2"
      onSubmit={handleSubmit}
    >
      <div className="min-w-72 flex-1">
        <EmailPrefixInput
          id="admin-user-email"
          label="邮箱"
          value={emailPrefix}
          onChange={setEmailPrefix}
          placeholder="jAccount"
        />
      </div>
      <Button type="submit">
        <RiSearchLine />
        查询
      </Button>
    </form>
  )
}

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

export function AdminUserQueryTab({
  currentUserID,
  currentUserIsSuperAdmin,
}: AdminUserQueryTabProps) {
  const search = routeApi.useSearch()
  const navigate = useNavigate({ from: "/admin/user" })
  const email = search.email ?? ""
  const page = Math.max(1, search.page ?? 1)
  const [suspendDays, setSuspendDays] = useState(30)
  const [suspendDialogOpen, setSuspendDialogOpen] = useState(false)
  const [resetDialogOpen, setResetDialogOpen] = useState(false)
  const [resetPassword, setResetPassword] = useState("")

  const userQuery = useAdminUserByEmail(email)
  const selectedUser = userQuery.data
  const reviewsQuery = useUserReviews(selectedUser?.id ?? 0, {
    page,
    page_size: reviewPageSize,
    order_by: "created_at",
  })
  const pointsQuery = useUserPoints(selectedUser?.id ?? 0, {
    page: 1,
    page_size: pointPageSize,
  })
  const suspendMutation = useSuspendAdminUser()
  const clearSuspensionMutation = useClearAdminUserSuspension()
  const grantAdminMutation = useGrantAdminUser()
  const revokeAdminMutation = useRevokeAdminUser()
  const resetPasswordMutation = useResetAdminUserPassword()

  function handleSearch(nextEmail: string) {
    void navigate({
      search: (prev) => ({
        ...prev,
        email: nextEmail || undefined,
        page: nextEmail ? 1 : undefined,
        tab: "user",
      }),
      replace: true,
      resetScroll: false,
    })
  }

  function setPage(nextPage: number) {
    void navigate({
      search: (prev) => ({ ...prev, page: nextPage, tab: "user" }),
      replace: true,
      resetScroll: false,
    })
  }

  async function clearSuspension() {
    if (!selectedUser) return
    await clearSuspensionMutation.mutateAsync(selectedUser.id)
  }

  async function confirmSuspension(event: FormSubmitEvent) {
    event.preventDefault()
    if (!selectedUser) return

    await suspendMutation.mutateAsync({
      userID: selectedUser.id,
      cmd: { days: suspendDays },
    })
    setSuspendDialogOpen(false)
  }

  async function grantAdmin() {
    if (!selectedUser) return
    await grantAdminMutation.mutateAsync(selectedUser.id)
  }

  async function revokeAdmin(userID: number) {
    await revokeAdminMutation.mutateAsync(userID)
  }

  function openResetPasswordDialog() {
    if (!selectedUser) return
    setResetPassword(
      defaultPasswordFromEmail(selectedUser.email, selectedUser.username)
    )
    setResetDialogOpen(true)
  }

  async function confirmResetPassword(event: FormSubmitEvent) {
    event.preventDefault()
    if (!selectedUser) return

    await resetPasswordMutation.mutateAsync({
      userID: selectedUser.id,
      cmd: { password: resetPassword },
    })
    setResetDialogOpen(false)
  }

  const isMutating =
    suspendMutation.isPending ||
    clearSuspensionMutation.isPending ||
    grantAdminMutation.isPending ||
    revokeAdminMutation.isPending ||
    resetPasswordMutation.isPending
  const selectedUserIsAdmin = selectedUser?.is_admin() ?? false
  const selectedUserIsSuperAdmin = selectedUser?.is_super_admin() ?? false
  const selectedUserIsSelf = selectedUser?.id === currentUserID

  return (
    <section className="space-y-6">
      <h2 className="text-lg font-medium">邮箱</h2>

      <AdminUserEmailSearchForm
        key={email}
        email={email}
        onSearch={handleSearch}
      />

      {userQuery.isLoading && email ? (
        <Skeleton className="h-36 w-full" />
      ) : null}

      {userQuery.isError && email ? (
        <p className="text-sm text-destructive">
          {getErrorMessage(userQuery.error)}
        </p>
      ) : null}

      {selectedUser ? (
        <section className="space-y-4">
          <div className="flex flex-wrap items-start justify-between gap-4 border-y py-4">
            <div className="grid gap-x-8 gap-y-3 text-sm sm:grid-cols-2 lg:grid-cols-3">
              <div>
                <p className="text-muted-foreground">用户 ID</p>
                <p className="font-mono">{selectedUser.id}</p>
              </div>
              <div>
                <p className="text-muted-foreground">用户名</p>
                <p className="font-mono">{selectedUser.username}</p>
              </div>
              <div>
                <p className="text-muted-foreground">邮箱</p>
                <p>{selectedUser.email}</p>
              </div>
              <div>
                <p className="text-muted-foreground">角色</p>
                <p>{selectedUser.role}</p>
              </div>
              <div>
                <p className="text-muted-foreground">注册时间</p>
                <p>{formatDateTime(selectedUser.created_at)}</p>
              </div>
              <div>
                <p className="text-muted-foreground">活跃时间</p>
                <p>{formatDateTime(selectedUser.last_seen_at)}</p>
              </div>
              <div className="sm:col-span-2 lg:col-span-3">
                <p className="text-muted-foreground">密码</p>
                <div className="mt-1 flex flex-wrap items-center gap-2">
                  <Badge
                    variant={
                      selectedUser.password_hash ? "secondary" : "outline"
                    }
                  >
                    {selectedUser.password_hash ? "已设置" : "未设置"}
                  </Badge>
                  {selectedUser.password_hash ? (
                    <span className="max-w-full font-mono text-xs break-all text-muted-foreground">
                      {selectedUser.password_hash}
                    </span>
                  ) : null}
                </div>
              </div>
              <div className="sm:col-span-2 lg:col-span-3">
                <p className="text-muted-foreground">封禁状态</p>
                <div className="mt-1 flex flex-wrap items-center gap-2">
                  <Badge
                    variant={
                      selectedUser.suspended ? "destructive" : "secondary"
                    }
                  >
                    {selectedUser.suspended ? "已封禁" : "正常"}
                  </Badge>
                  <SuspensionText
                    suspendedAt={selectedUser.suspended_at}
                    suspendTill={selectedUser.suspend_till}
                  />
                </div>
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-end gap-2">
              {selectedUserIsSelf ? (
                <p className="text-sm text-muted-foreground">不能对自己操作</p>
              ) : selectedUserIsSuperAdmin ? (
                <p className="text-sm text-muted-foreground">
                  不能修改超级管理员权限
                </p>
              ) : !currentUserIsSuperAdmin ? (
                <p className="text-sm text-muted-foreground">
                  需要超级管理员权限
                </p>
              ) : selectedUserIsAdmin ? (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      className="hover:border-destructive/40 hover:bg-destructive/10 hover:text-destructive focus-visible:border-destructive/40 focus-visible:ring-destructive/20 dark:hover:bg-destructive/20"
                      disabled={isMutating}
                    >
                      <RiShieldCrossLine />
                      撤销权限
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent size="sm">
                    <AlertDialogHeader>
                      <AlertDialogTitle>撤销管理员权限</AlertDialogTitle>
                      <AlertDialogDescription>
                        撤销 {selectedUser.username} 的管理员权限。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction
                        variant="destructive"
                        onClick={() => {
                          void revokeAdmin(selectedUser.id)
                        }}
                      >
                        撤销
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              ) : (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button type="button" variant="outline" disabled={isMutating}>
                      <RiShieldUserLine />
                      授予 admin
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent size="sm">
                    <AlertDialogHeader>
                      <AlertDialogTitle>授予管理员权限</AlertDialogTitle>
                      <AlertDialogDescription>
                        授予 {selectedUser.username} 管理员权限。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => {
                          void grantAdmin()
                        }}
                      >
                        授予
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              )}

              {currentUserIsSuperAdmin && !selectedUserIsSelf ? (
                <Dialog
                  open={resetDialogOpen}
                  onOpenChange={setResetDialogOpen}
                >
                  <DialogTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={openResetPasswordDialog}
                      disabled={isMutating}
                    >
                      <RiKey2Line />
                      重置密码
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <form onSubmit={confirmResetPassword} className="space-y-6">
                      <DialogHeader>
                        <DialogTitle>重置密码</DialogTitle>
                        <DialogDescription>
                          用户 {selectedUser.id} 的密码会被立即改为下方内容。
                        </DialogDescription>
                      </DialogHeader>

                      <div className="space-y-2">
                        <Label htmlFor="reset-password">新密码</Label>
                        <Input
                          id="reset-password"
                          type="text"
                          value={resetPassword}
                          onChange={(event) =>
                            setResetPassword(event.target.value)
                          }
                          autoFocus
                        />
                      </div>

                      {resetPasswordMutation.isError ? (
                        <p className="text-sm text-destructive">
                          {getErrorMessage(resetPasswordMutation.error)}
                        </p>
                      ) : null}

                      <DialogFooter>
                        <DialogClose asChild>
                          <Button type="button" variant="outline">
                            取消
                          </Button>
                        </DialogClose>
                        <Button
                          type="submit"
                          disabled={
                            resetPasswordMutation.isPending ||
                            resetPassword.trim() === ""
                          }
                        >
                          确认重置
                        </Button>
                      </DialogFooter>
                    </form>
                  </DialogContent>
                </Dialog>
              ) : null}

              {!selectedUserIsSelf &&
                (selectedUserIsAdmin ? (
                  <p className="text-sm text-muted-foreground">
                    管理员不能被封禁
                  </p>
                ) : selectedUser.suspended ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={clearSuspension}
                    disabled={isMutating}
                  >
                    <RiLockUnlockLine />
                    解封用户
                  </Button>
                ) : (
                  <Dialog
                    open={suspendDialogOpen}
                    onOpenChange={setSuspendDialogOpen}
                  >
                    <DialogTrigger asChild>
                      <Button
                        type="button"
                        variant="destructive"
                        disabled={isMutating}
                      >
                        <RiLockLine />
                        封禁用户
                      </Button>
                    </DialogTrigger>
                    <DialogContent>
                      <form onSubmit={confirmSuspension} className="space-y-6">
                        <DialogHeader>
                          <DialogTitle>封禁用户</DialogTitle>
                          <DialogDescription>
                            用户 {selectedUser.id}{" "}
                            将在封禁期间无法登录或继续操作。
                          </DialogDescription>
                        </DialogHeader>

                        <div className="space-y-2">
                          <Label htmlFor="suspend-days">封禁天数</Label>
                          <Input
                            id="suspend-days"
                            type="number"
                            min={1}
                            value={suspendDays}
                            onChange={(event) =>
                              setSuspendDays(
                                Math.max(1, Number(event.target.value) || 30)
                              )
                            }
                            autoFocus
                          />
                        </div>

                        {suspendMutation.isError ? (
                          <p className="text-sm text-destructive">
                            {getErrorMessage(suspendMutation.error)}
                          </p>
                        ) : null}

                        <DialogFooter>
                          <DialogClose asChild>
                            <Button type="button" variant="outline">
                              取消
                            </Button>
                          </DialogClose>
                          <Button
                            type="submit"
                            variant="destructive"
                            disabled={suspendMutation.isPending}
                          >
                            确认封禁
                          </Button>
                        </DialogFooter>
                      </form>
                    </DialogContent>
                  </Dialog>
                ))}
            </div>
          </div>

          <div className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-medium">积分记录</h2>
                {pointsQuery.data ? (
                  <p className="mt-1 text-sm text-muted-foreground">
                    当前 {pointsQuery.data.total} 分，共{" "}
                    {pointsQuery.data.records.total} 条记录
                  </p>
                ) : null}
              </div>
            </div>

            {pointsQuery.isLoading ? (
              <div className="space-y-3">
                {[...Array(5)].map((_, i) => (
                  <Skeleton key={i} className="h-14 w-full" />
                ))}
              </div>
            ) : pointsQuery.isError ? (
              <p className="text-sm text-destructive">
                {getErrorMessage(pointsQuery.error)}
              </p>
            ) : (
              <PointRecordList
                records={pointsQuery.data?.records.items ?? []}
              />
            )}
          </div>

          <Separator />

          <div className="space-y-3">
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-lg font-medium">点评记录</h2>
              {reviewsQuery.data ? (
                <p className="text-sm text-muted-foreground">
                  共 {reviewsQuery.data.total} 条
                </p>
              ) : null}
            </div>

            <ReviewList
              reviews={reviewsQuery.data?.items ?? []}
              isLoading={reviewsQuery.isLoading}
              showCourse
              emptyText="该用户暂无点评"
            />

            {reviewsQuery.data && reviewsQuery.data.total > reviewPageSize ? (
              <PaginationComponent
                page={reviewsQuery.data.page}
                pageSize={reviewsQuery.data.page_size}
                total={reviewsQuery.data.total}
                onPageChange={setPage}
              />
            ) : null}
          </div>
        </section>
      ) : null}
    </section>
  )
}
