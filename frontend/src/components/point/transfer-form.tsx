import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useCreateTransfer, usePreviewTransfer } from "@/hooks/use-point"
import type { FeePayer } from "@/api/point"

interface TransferFormProps {
  onSuccess?: () => void
}

export function TransferForm({ onSuccess }: TransferFormProps) {
  const previewMutation = usePreviewTransfer()
  const transferMutation = useCreateTransfer()
  const [recipient, setRecipient] = useState("")
  const [amount, setAmount] = useState("")
  const [feePayer, setFeePayer] = useState<FeePayer>("sender")
  const [preview, setPreview] = useState<{
    fee: number
    sender_debit: number
    recipient_credit: number
  } | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function handlePreview() {
    setError(null)
    const num = Number(amount)
    if (!Number.isFinite(num) || num <= 0) {
      setError("请输入有效金额")
      return
    }
    try {
      const result = await previewMutation.mutateAsync({
        recipient_username: recipient,
        amount: num,
        fee_payer: feePayer,
      })
      setPreview({ fee: result.fee, sender_debit: result.sender_debit, recipient_credit: result.recipient_credit })
    } catch (err) {
      setError(err instanceof Error ? err.message : "预览失败")
    }
  }

  async function handleSubmit() {
    setError(null)
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

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="recipient">收款人用户名</Label>
        <Input
          id="recipient"
          placeholder="输入用户名"
          value={recipient}
          onChange={(e) => { setRecipient(e.target.value); setPreview(null) }}
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="amount">转账金额</Label>
          <Input
            id="amount"
            type="number"
            placeholder="0"
            value={amount}
            onChange={(e) => { setAmount(e.target.value); setPreview(null) }}
            min={1}
          />
        </div>
        <div className="space-y-2">
          <Label>手续费承担</Label>
          <Select
            value={feePayer}
            onValueChange={(v) => { setFeePayer(v as FeePayer); setPreview(null) }}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="sender">我承担</SelectItem>
              <SelectItem value="recipient">对方承担</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {preview && (
        <div className="rounded-md border p-3 space-y-1 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">手续费</span>
            <span>{preview.fee}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">你的支出</span>
            <span className="text-destructive">-{preview.sender_debit}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">对方收入</span>
            <span className="text-green-600">+{preview.recipient_credit}</span>
          </div>
        </div>
      )}

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex gap-2">
        <Button
          type="button"
          variant="outline"
          className="flex-1"
          onClick={handlePreview}
          disabled={isLoading || !recipient || !amount}
        >
          {previewMutation.isPending ? "计算中..." : "预览"}
        </Button>
        <Button
          type="button"
          className="flex-1"
          onClick={handleSubmit}
          disabled={isLoading || !preview}
        >
          {transferMutation.isPending ? "转账中..." : "确认转账"}
        </Button>
      </div>
    </div>
  )
}
