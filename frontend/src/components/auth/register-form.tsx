import { useState } from "react"
import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { CodeInputWithButton } from "./code-button"
import { useAuth } from "@/contexts/auth-context"

export function RegisterForm() {
  const { register, sendRegisterCode } = useAuth()
  const [email, setEmail] = useState("")
  const [code, setCode] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!email || !code || !password) {
      setError("请填写所有字段")
      return
    }
    if (password.length < 6) {
      setError("密码至少 6 位")
      return
    }
    setIsLoading(true)
    try {
      await register({ email, code, password })
    } catch (err) {
      setError(err instanceof Error ? err.message : "注册失败")
      setIsLoading(false)
    }
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>注册</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">邮箱（仅支持 @sjtu.edu.cn）</Label>
            <Input
              id="email"
              type="email"
              placeholder="your@sjtu.edu.cn"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
            />
          </div>
          <CodeInputWithButton
            id="code"
            label="验证码"
            email={email}
            code={code}
            onCodeChange={setCode}
            onSend={sendRegisterCode}
          />
          <div className="space-y-2">
            <Label htmlFor="password">密码</Label>
            <Input
              id="password"
              type="password"
              placeholder="至少 6 位"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "注册中..." : "注册"}
          </Button>
          <p className="text-center text-sm text-muted-foreground">
            已有账号？
            <Link to="/login" className="ml-1 text-foreground hover:underline">
              登录
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  )
}
