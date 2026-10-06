import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useAuth } from "@/contexts/auth-context"
import {
  getAccountBindingURL,
  updateAccountProfile,
  unbindAccount,
} from "@/api/auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PasswordForm } from "./password-form"

export function AccountSettings() {
  const { user } = useAuth()
  const query = useQueryClient()
  const [message, setMessage] = useState("")
  const [busy, setBusy] = useState(false)
  if (!user) return null
  async function save(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    setBusy(true)
    setMessage("")
    try {
      await updateAccountProfile(
        (data.get("username") as string | null) ?? "",
        (data.get("nickname") as string | null) ?? ""
      )
      await query.invalidateQueries({ queryKey: ["auth", "me"] })
      setMessage("资料已保存")
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "保存失败")
    } finally {
      setBusy(false)
    }
  }
  async function unbind() {
    setBusy(true)
    setMessage("")
    try {
      await unbindAccount()
      window.location.assign("/login")
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "解除绑定失败")
      setBusy(false)
    }
  }
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>账号资料</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={save}>
            <div className="space-y-2">
              <Label htmlFor="profile-username">本站账号名</Label>
              <Input
                id="profile-username"
                name="username"
                defaultValue={user.username}
                required
                minLength={3}
                maxLength={32}
              />
              <p className="text-xs text-muted-foreground">
                用于本站账密登录，可使用小写字母、数字、点、横线或下划线。
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-nickname">昵称</Label>
              <Input
                id="profile-nickname"
                name="nickname"
                defaultValue={
                  user.nickname || user.bound_nickname || user.username
                }
                required
                maxLength={64}
              />
            </div>
            <Button disabled={busy} type="submit">
              保存资料
            </Button>
          </form>
          {message && (
            <p role="status" className="mt-3 text-sm">
              {message}
            </p>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>资料站账号</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {user.subject
              ? `已绑定：${user.bound_nickname || "资料站账号"}`
              : "尚未绑定。绑定后可通过资料站登录，并使用资料站积分。"}
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              onClick={() => window.location.assign(getAccountBindingURL())}
            >
              {user.subject ? "更换绑定账号" : "绑定资料站账号"}
            </Button>
            {user.subject && (
              <Button
                variant="outline"
                disabled={busy || !user.has_password}
                onClick={() => void unbind()}
              >
                解除绑定
              </Button>
            )}
          </div>
          {user.subject && !user.has_password && (
            <p className="text-xs text-muted-foreground">
              先设置本站密码，再解除绑定，以便继续登录。
            </p>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>
            {user.has_password ? "修改本站密码" : "设置本站密码"}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <PasswordForm />
        </CardContent>
      </Card>
    </>
  )
}
