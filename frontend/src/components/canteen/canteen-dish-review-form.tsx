import { useState } from "react"

import type { CanteenMediaAssetDTO } from "@/api/canteen"
import { CanteenMediaUploader } from "@/components/canteen/canteen-media-uploader"
import { CanteenRatingInput } from "@/components/canteen/canteen-rating-input"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { useCreateCanteenDishReview } from "@/hooks/use-canteen-dishes"

export function CanteenDishReviewForm({
  dishID,
  onDone,
}: {
  dishID: number
  onDone?: () => void
}) {
  const [taste, setTaste] = useState(4)
  const [portion, setPortion] = useState(4)
  const [value, setValue] = useState(4)
  const [content, setContent] = useState("")
  const [mediaAssets, setMediaAssets] = useState<CanteenMediaAssetDTO[]>([])
  const [error, setError] = useState<string | null>(null)
  const create = useCreateCanteenDishReview(dishID)

  async function submit() {
    setError(null)
    try {
      await create.mutateAsync({
        taste,
        portion,
        value,
        content,
        media_ids: mediaAssets.map((asset) => asset.id),
      })
      setContent("")
      setMediaAssets([])
      onDone?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : "提交菜品评分失败")
    }
  }

  return (
    <div className="space-y-3 rounded-md border p-3">
      <CanteenRatingInput label="口味" value={taste} onChange={setTaste} />
      <CanteenRatingInput label="分量" value={portion} onChange={setPortion} />
      <CanteenRatingInput label="性价比" value={value} onChange={setValue} />
      <Textarea
        rows={2}
        maxLength={500}
        placeholder="这道菜怎么样？（可选）"
        value={content}
        onChange={(event) => setContent(event.target.value)}
      />
      <CanteenMediaUploader
        value={mediaAssets}
        onChange={setMediaAssets}
        purpose="dish"
      />
      {error && <p className="text-sm text-destructive">{error}</p>}
      <Button
        type="button"
        size="sm"
        disabled={create.isPending}
        onClick={() => void submit()}
      >
        {create.isPending ? "提交中…" : "提交菜品评分"}
      </Button>
    </div>
  )
}
