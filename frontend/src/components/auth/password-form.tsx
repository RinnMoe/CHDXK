import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { changePassword } from "@/api/auth"
import { useAuth } from "@/contexts/auth-context"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function PasswordForm({ required = false }: { required?: boolean }) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState("")
  const [error, setError] = useState("")
  async function submit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = event.currentTarget
    const data = new FormData(form)
    const password = (data.get("password") as string | null) ?? ""
    if (password !== data.get("confirm")) {
      setError("两次输入的密码不一致")
      return
    }
    setBusy(true)
    setError("")
    setMessage("")
    try {
      await changePassword(
        (data.get("old_password") as string | null) ?? "",
        password
      )
      form.reset()
      await queryClient.invalidateQueries({ queryKey: ["auth", "me"] })
      if (required) window.location.replace("/")
      else setMessage("密码已更新，其他设备需重新登录。")
    } catch (err) {
      setError(err instanceof Error ? err.message : "密码更新失败")
    } finally {
      setBusy(false)
    }
  }
  return (
    <form className="space-y-4" onSubmit={(event) => void submit(event)}>
      {user?.has_password && (
        <div className="space-y-2">
          <Label htmlFor="old-password">原密码</Label>
          <Input
            id="old-password"
            name="old_password"
            type="password"
            autoComplete="current-password"
            required
          />
        </div>
      )}
      <div className="space-y-2">
        <Label htmlFor="new-password">新密码</Label>
        <Input
          id="new-password"
          name="password"
          type="password"
          autoComplete="new-password"
          minLength={8}
          required
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="confirm-password">确认新密码</Label>
        <Input
          id="confirm-password"
          name="confirm"
          type="password"
          autoComplete="new-password"
          minLength={8}
          required
        />
      </div>
      <p className="text-xs text-muted-foreground">
        至少 8 个字符。请使用与资料站不同的密码。
      </p>
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
      {message && (
        <p role="status" className="text-sm text-muted-foreground">
          {message}
        </p>
      )}
      <Button type="submit" disabled={busy}>
        {busy ? "保存中…" : "保存密码"}
      </Button>
    </form>
  )
}
