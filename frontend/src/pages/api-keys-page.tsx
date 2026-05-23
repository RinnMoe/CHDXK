import { useState, type FormEvent } from "react"
import { Navigate } from "react-router-dom"
import { RiAddLine, RiDeleteBinLine, RiFileCopyLine } from "@remixicon/react"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
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
import { useAuth } from "@/contexts/auth-context"
import {
  useApiKeys,
  useCreateApiKey,
  useDeleteApiKey,
} from "@/hooks/use-api-key"
import { formatNullableDateTime } from "@/lib/date"
import type { ApiKeyDTO } from "@/api/api-key"

export function ApiKeysPage() {
  const { user, isLoading: authLoading } = useAuth()
  const { data: apiKeys = [], isLoading } = useApiKeys(!!user)
  const createMutation = useCreateApiKey()
  const deleteMutation = useDeleteApiKey()

  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [name, setName] = useState("")
  const [createdKey, setCreatedKey] = useState<ApiKeyDTO | null>(null)
  const [copyMessage, setCopyMessage] = useState("")
  const [error, setError] = useState("")

  if (authLoading) return null
  if (!user) return <Navigate to="/login" replace />

  async function copyKey(value: string) {
    setCopyMessage("")
    try {
      await navigator.clipboard.writeText(value)
      setCopyMessage("已复制")
    } catch {
      setCopyMessage("复制失败")
    }
  }

  function handleCreateOpenChange(open: boolean) {
    setIsCreateOpen(open)
    if (!open) {
      setName("")
      setCreatedKey(null)
      setCopyMessage("")
      setError("")
    }
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = name.trim()
    if (!trimmed) {
      setError("请输入名称")
      return
    }
    setError("")
    setCopyMessage("")
    try {
      const key = await createMutation.mutateAsync({ name: trimmed })
      setCreatedKey(key)
      setName("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建失败")
    }
  }

  async function handleDelete(id: number) {
    setError("")
    try {
      await deleteMutation.mutateAsync(id)
      if (createdKey?.id === id) setCreatedKey(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败")
    }
  }

  return (
    <>
      <PageTitle>API Keys</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h1 className="text-2xl font-semibold">API Keys</h1>
            <Button onClick={() => handleCreateOpenChange(true)}>
              <RiAddLine data-icon="inline-start" />
              新建 API Key
            </Button>
          </div>

          {error && !isCreateOpen && (
            <p className="text-sm text-destructive">{error}</p>
          )}

          <Dialog open={isCreateOpen} onOpenChange={handleCreateOpenChange}>
            <DialogContent className="sm:max-w-2xl">
              <DialogHeader>
                <DialogTitle>新建 API Key</DialogTitle>
                <DialogDescription>
                  创建成功后明文仅显示一次。
                </DialogDescription>
              </DialogHeader>

              {createdKey?.key ? (
                <div className="space-y-4">
                  <div className="flex flex-wrap items-start gap-3">
                    <code className="min-w-0 flex-1 rounded-md border bg-muted/30 px-3 py-2 text-sm break-all">
                      {createdKey.key}
                    </code>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => copyKey(createdKey.key!)}
                    >
                      <RiFileCopyLine data-icon="inline-start" />
                      复制
                    </Button>
                  </div>
                  {copyMessage && (
                    <p className="text-sm text-muted-foreground">
                      {copyMessage}
                    </p>
                  )}
                  <DialogFooter>
                    <DialogClose asChild>
                      <Button>完成</Button>
                    </DialogClose>
                  </DialogFooter>
                </div>
              ) : (
                <form className="space-y-4" onSubmit={handleCreate}>
                  <div className="space-y-2">
                    <Label htmlFor="api-key-name">名称</Label>
                    <Input
                      id="api-key-name"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      placeholder="例如：本地脚本"
                      maxLength={80}
                    />
                  </div>
                  {error && <p className="text-sm text-destructive">{error}</p>}
                  <DialogFooter>
                    <DialogClose asChild>
                      <Button type="button" variant="outline">
                        取消
                      </Button>
                    </DialogClose>
                    <Button type="submit" disabled={createMutation.isPending}>
                      {createMutation.isPending ? "创建中" : "创建"}
                    </Button>
                  </DialogFooter>
                </form>
              )}
            </DialogContent>
          </Dialog>

          <Card>
            <CardHeader>
              <CardTitle>我的 API Keys</CardTitle>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="space-y-3">
                  {[...Array(4)].map((_, i) => (
                    <Skeleton key={i} className="h-10 w-full" />
                  ))}
                </div>
              ) : apiKeys.length === 0 ? (
                <p className="py-8 text-center text-sm text-muted-foreground">
                  暂无 API Key
                </p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>名称</TableHead>
                      <TableHead>Key</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>最后使用</TableHead>
                      <TableHead className="w-16 text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {apiKeys.map((key) => (
                      <TableRow key={key.id}>
                        <TableCell className="font-medium">
                          {key.name}
                        </TableCell>
                        <TableCell>
                          <code className="rounded bg-muted px-2 py-1 text-xs">
                            {key.key_masked}
                          </code>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatNullableDateTime(key.created_at)}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatNullableDateTime(
                            key.last_used_at,
                            "从未使用"
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button
                                variant="destructive"
                                size="icon-sm"
                                aria-label="删除 API Key"
                              >
                                <RiDeleteBinLine />
                              </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent size="sm">
                              <AlertDialogHeader>
                                <AlertDialogTitle>
                                  删除 API Key
                                </AlertDialogTitle>
                                <AlertDialogDescription>
                                  删除后使用该 key 的请求会立即失效。
                                </AlertDialogDescription>
                              </AlertDialogHeader>
                              <AlertDialogFooter>
                                <AlertDialogCancel>取消</AlertDialogCancel>
                                <AlertDialogAction
                                  variant="destructive"
                                  onClick={() => handleDelete(key.id)}
                                >
                                  删除
                                </AlertDialogAction>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      </PageShell>
    </>
  )
}
