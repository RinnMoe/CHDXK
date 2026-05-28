import { useState } from "react"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { RiLockLine, RiLockUnlockLine, RiSearchLine } from "@remixicon/react"
import { PaginationComponent } from "@/components/common/pagination"
import { ReviewList } from "@/components/review/review-list"
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
import { Skeleton } from "@/components/ui/skeleton"
import {
  useAdminUserByEmail,
  useClearAdminUserSuspension,
  useGrantAdminUser,
  useRevokeAdminUser,
  useSuspendAdminUser,
} from "@/hooks/use-admin-user"
import { useUserReviews } from "@/hooks/use-review"
import { formatDateTime, formatNullableDateTime } from "@/lib/date"
import { getErrorMessage, type FormSubmitEvent } from "./admin-utils"

const reviewPageSize = 20
const routeApi = getRouteApi("/app/admin/user")

interface AdminUserQueryTabProps {
  currentUserID: number
  currentUserIsSuperAdmin: boolean
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

  const userQuery = useAdminUserByEmail(email)
  const selectedUser = userQuery.data
  const reviewsQuery = useUserReviews(selectedUser?.id ?? 0, {
    page,
    page_size: reviewPageSize,
    order_by: "created_at",
  })
  const suspendMutation = useSuspendAdminUser()
  const clearSuspensionMutation = useClearAdminUserSuspension()
  const grantAdminMutation = useGrantAdminUser()
  const revokeAdminMutation = useRevokeAdminUser()

  function handleSearch(event: FormSubmitEvent) {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const nextEmail = String(formData.get("email") ?? "")
      .trim()
      .toLowerCase()
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

  const isMutating =
    suspendMutation.isPending ||
    clearSuspensionMutation.isPending ||
    grantAdminMutation.isPending ||
    revokeAdminMutation.isPending
  const selectedUserIsAdmin = selectedUser?.is_admin() ?? false
  const selectedUserIsSuperAdmin = selectedUser?.is_super_admin() ?? false
  const selectedUserIsSelf = selectedUser?.id === currentUserID

  return (
    <section className="space-y-6">
      <h2 className="text-lg font-medium">邮箱</h2>

      <form
        key={email}
        className="flex flex-wrap items-end gap-2"
        onSubmit={handleSearch}
      >
        <div className="min-w-72 flex-1 space-y-1">
          <Label htmlFor="admin-user-email" className="sr-only">
            邮箱
          </Label>
          <Input
            id="admin-user-email"
            name="email"
            type="email"
            defaultValue={email}
            placeholder="name@example.edu"
          />
        </div>
        <Button type="submit">
          <RiSearchLine />
          查询
        </Button>
      </form>

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
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => revokeAdmin(selectedUser.id)}
                  disabled={isMutating}
                >
                  撤销权限
                </Button>
              ) : (
                <Button
                  type="button"
                  variant="outline"
                  onClick={grantAdmin}
                  disabled={isMutating}
                >
                  授予 admin
                </Button>
              )}

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
