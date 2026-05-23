import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { EmailPrefixInput } from "@/components/auth/email-prefix-input"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { CodeInputWithButton } from "./code-button"
import { buildAuthEmail, validateAuthPassword } from "@/config/auth"
import { useAuth } from "@/contexts/auth-context"

export function PasswordResetForm() {
  const navigate = useNavigate()
  const { resetPassword, sendResetCode } = useAuth()
  const [emailPrefix, setEmailPrefix] = useState("")
  const [code, setCode] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const email = emailPrefix.trim() ? buildAuthEmail(emailPrefix) : ""

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!emailPrefix.trim() || !code || !password || !confirmPassword) {
      setError("请填写所有字段")
      return
    }
    const passwordError = validateAuthPassword(password)
    if (passwordError) {
      setError(passwordError)
      return
    }
    if (password !== confirmPassword) {
      setError("两次输入的密码不一致")
      return
    }
    setIsLoading(true)
    try {
      await resetPassword({ email, code, new_password: password })
      navigate("/login")
    } catch (err) {
      setError(err instanceof Error ? err.message : "重置失败")
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>重置密码</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <EmailPrefixInput
            id="email"
            label="邮箱"
            value={emailPrefix}
            onChange={setEmailPrefix}
          />
          <CodeInputWithButton
            id="code"
            label="验证码"
            email={email}
            code={code}
            onCodeChange={setCode}
            onSend={sendResetCode}
          />
          <div className="space-y-2">
            <Label htmlFor="password">新密码</Label>
            <Input
              id="password"
              type="password"
              placeholder="至少 10 位，含字母/数字/符号"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirm-password">确认密码</Label>
            <Input
              id="confirm-password"
              type="password"
              placeholder="再次输入密码"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "重置中..." : "重置密码"}
          </Button>
          <p className="text-center text-sm text-muted-foreground">
            已想起密码？
            <Link to="/login" className="ml-1 text-foreground hover:underline">
              返回登录
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  )
}
