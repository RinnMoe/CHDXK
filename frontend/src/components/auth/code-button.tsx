import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface CodeButtonProps {
  email: string
  onSend: (email: string) => Promise<void>
  disabled?: boolean
}

const COUNTDOWN_SECONDS = 60

export function CodeButton({ email, onSend, disabled }: CodeButtonProps) {
  const [seconds, setSeconds] = useState(0)
  const [isSending, setIsSending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (seconds <= 0) return
    const timer = setInterval(() => {
      setSeconds((s) => Math.max(0, s - 1))
    }, 1000)
    return () => clearInterval(timer)
  }, [seconds])

  async function handleClick() {
    setError(null)
    if (!email) {
      setError("请先填写邮箱")
      return
    }
    setIsSending(true)
    try {
      await onSend(email)
      setSeconds(COUNTDOWN_SECONDS)
    } catch (err) {
      setError(err instanceof Error ? err.message : "发送失败")
    } finally {
      setIsSending(false)
    }
  }

  return (
    <div className="space-y-1">
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="w-full"
        onClick={handleClick}
        disabled={disabled || isSending || seconds > 0}
      >
        {seconds > 0
          ? `${seconds} 秒后可重发`
          : isSending
            ? "发送中..."
            : "发送验证码"}
      </Button>
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}

interface CodeInputProps {
  id: string
  label: string
  email: string
  code: string
  onCodeChange: (v: string) => void
  onSend: (email: string) => Promise<void>
  disabled?: boolean
}

export function CodeInputWithButton({
  id,
  label,
  email,
  code,
  onCodeChange,
  onSend,
  disabled,
}: CodeInputProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="grid grid-cols-[1fr_auto] gap-2 items-start">
        <Input
          id={id}
          placeholder="6 位验证码"
          value={code}
          onChange={(e) => onCodeChange(e.target.value)}
          inputMode="numeric"
          maxLength={6}
          autoComplete="one-time-code"
        />
        <CodeButton email={email} onSend={onSend} disabled={disabled} />
      </div>
    </div>
  )
}
