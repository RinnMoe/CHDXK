import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { useAuth } from "@/contexts/auth-context"
import { useCreateTransfer, usePreviewTransfer, useUserPoints } from "@/hooks/use-point"
import type { FeePayer } from "@/api/point"

interface TransferFormProps {
  onSuccess?: () => void
}

export function TransferForm({ onSuccess }: TransferFormProps) {
  const { user } = useAuth()
  const previewMutation = usePreviewTransfer()
  const transferMutation = useCreateTransfer()
  const { data: pointsData } = useUserPoints(user?.id ?? 0, { page: 1, page_size: 1 })

  const [recipient, setRecipient] = useState("")
  const [amount, setAmount] = useState("")
  const [feePayer, setFeePayer] = useState<FeePayer>("sender")
  const [preview, setPreview] = useState<{
    fee: number
    sender_debit: number
    recipient_credit: number
    sender_remaining: number
  } | null>(null)
  const [error, setError] = useState<string | null>(null)

  const previewMutate = previewMutation.mutateAsync
  useEffect(() => {
    const num = Number(amount)
    if (!amount || !Number.isFinite(num) || num <= 0) {
      setPreview(null)
      return
    }
    const handler = setTimeout(async () => {
      setError(null)
      try {
        const result = await previewMutate({
          amount: num,
          fee_payer: feePayer,
        })
        setPreview({
          fee: result.fee,
          sender_debit: result.sender_debit,
          recipient_credit: result.recipient_credit,
          sender_remaining: result.sender_remaining,
        })
      } catch (err) {
        setPreview(null)
        setError(err instanceof Error ? err.message : "预览失败")
      }
    }, 200)
    return () => clearTimeout(handler)
  }, [amount, feePayer, previewMutate])

  async function handleSubmit() {
    setError(null)
    if (!recipient) {
      setError("请输入收款人用户名")
      return
    }
    const num = Number(amount)
    if (!Number.isFinite(num) || num <= 0) {
      setError("请输入有效金额")
      return
    }
    try {
      await transferMutation.mutateAsync({
        recipient_username: recipient,
        amount: num,
        fee_payer: feePayer,
      })
      setPreview(null)
      setRecipient("")
      setAmount("")
      onSuccess?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : "转账失败")
    }
  }

  const isLoading = previewMutation.isPending || transferMutation.isPending
  const currentBalance = pointsData?.total ?? 0
  const insufficientBalance = preview != null && preview.sender_remaining < 0

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="recipient">收款人用户名</Label>
        <Input
          id="recipient"
          placeholder="输入用户名"
          value={recipient}
          onChange={(e) => setRecipient(e.target.value)}
        />
      </div>
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="amount">转账金额</Label>
          <span className="text-sm text-muted-foreground">剩余 {currentBalance} 积分</span>
        </div>
        <Input
          id="amount"
          type="number"
          placeholder="0"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          min={1}
        />
      </div>
      <div className="space-y-2">
        <Label>手续费承担</Label>
        <RadioGroup
          value={feePayer}
          onValueChange={(v) => setFeePayer(v as FeePayer)}
          className="flex gap-6"
        >
          <div className="flex items-center gap-2">
            <RadioGroupItem value="sender" id="fee-sender" />
            <Label htmlFor="fee-sender" className="font-normal">我承担</Label>
          </div>
          <div className="flex items-center gap-2">
            <RadioGroupItem value="recipient" id="fee-recipient" />
            <Label htmlFor="fee-recipient" className="font-normal">对方承担</Label>
          </div>
        </RadioGroup>
      </div>

      {preview && (
        <div className="rounded-md border p-3 space-y-1 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">当前积分</span>
            <span>{currentBalance}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">转账金额</span>
            <span>-{amount}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">手续费</span>
            <span>-{preview.fee}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">对方收入</span>
            <span className="text-green-600">+{preview.recipient_credit}</span>
          </div>
          <div className="flex justify-between font-medium">
            <span>转账后剩余</span>
            <span className={insufficientBalance ? "text-destructive" : ""}>
              {preview.sender_remaining}
            </span>
          </div>
          <p className="text-xs text-muted-foreground pt-1">* 实际结果以转账后为准</p>
        </div>
      )}

      {error && <p className="text-sm text-destructive">{error}</p>}

      {insufficientBalance && (
        <p className="text-sm text-destructive">积分不足，转账后剩余 {preview!.sender_remaining}</p>
      )}

      <Button
        type="button"
        className="w-full"
        onClick={handleSubmit}
        disabled={isLoading || !preview || !recipient || insufficientBalance}
      >
        {transferMutation.isPending ? "转账中..." : "确认转账"}
      </Button>
    </div>
  )
}
