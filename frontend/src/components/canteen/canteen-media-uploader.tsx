import { useState } from "react"
import { RiCloseLine } from "@remixicon/react"

import type { CanteenMediaAssetDTO } from "@/api/canteen"
import { canteenMediaURL } from "@/api/canteen-community"
import { useUploadCanteenMedia } from "@/hooks/use-canteen-media"
import { Button } from "@/components/ui/button"

const maxImages = 3

async function prepareImage(file: File): Promise<File> {
  if (file.type !== "image/jpeg" && file.type !== "image/png") {
    throw new Error("仅支持 JPEG/PNG 图片")
  }
  const bitmap = await createImageBitmap(file)
  const maxEdge = 2048
  const scale = Math.min(1, maxEdge / Math.max(bitmap.width, bitmap.height))
  if (scale === 1 && file.size <= 4 * 1024 * 1024) {
    return file
  }
  const canvas = document.createElement("canvas")
  canvas.width = Math.round(bitmap.width * scale)
  canvas.height = Math.round(bitmap.height * scale)
  const context = canvas.getContext("2d")
  if (!context) throw new Error("无法处理图片")
  context.drawImage(bitmap, 0, 0, canvas.width, canvas.height)
  const blob = await new Promise<Blob | null>((resolve) =>
    canvas.toBlob(resolve, "image/jpeg", 0.85)
  )
  if (!blob) throw new Error("图片压缩失败")
  return new File([blob], file.name.replace(/\.\w+$/, ".jpg"), {
    type: "image/jpeg",
  })
}

export function CanteenMediaUploader({
  value,
  onChange,
  purpose = "review",
}: {
  value: CanteenMediaAssetDTO[]
  onChange: (assets: CanteenMediaAssetDTO[]) => void
  purpose?: "review" | "dish"
}) {
  const upload = useUploadCanteenMedia(purpose)
  const [error, setError] = useState<string | null>(null)
  const [preparing, setPreparing] = useState(false)

  async function handleFiles(files: FileList | null) {
    if (!files || files.length === 0) return
    setError(null)
    setPreparing(true)
    const next = [...value]
    for (const file of Array.from(files).slice(0, maxImages - value.length)) {
      try {
        const prepared = await prepareImage(file)
        const asset = await upload.mutateAsync(prepared)
        next.push(asset)
      } catch (err) {
        setError(err instanceof Error ? err.message : "上传图片失败")
      }
    }
    setPreparing(false)
    onChange(next)
  }

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        {value.map((asset) => (
          <div key={asset.id} className="relative">
            <img
              src={canteenMediaURL(asset.id)}
              alt="实拍预览"
              className="h-20 w-20 rounded-md border object-cover"
            />
            <button
              type="button"
              aria-label="移除图片"
              className="absolute -top-2 -right-2 rounded-full bg-background p-0.5 shadow ring-1 ring-border"
              onClick={() => onChange(value.filter((item) => item.id !== asset.id))}
            >
              <RiCloseLine className="size-3.5" />
            </button>
          </div>
        ))}
        {value.length < maxImages && (
          <label className="flex h-20 w-20 cursor-pointer items-center justify-center rounded-md border border-dashed text-xs text-muted-foreground hover:bg-muted/40">
            <input
              type="file"
              accept="image/jpeg,image/png"
              multiple
              className="hidden"
              disabled={preparing || upload.isPending}
              onChange={(event) => {
                void handleFiles(event.target.files)
                event.target.value = ""
              }}
            />
            {preparing || upload.isPending ? "上传中…" : "添加图片"}
          </label>
        )}
      </div>
      <p className="text-xs text-muted-foreground">
        最多 3 张，上传后需管理员审核才会公开。图片会自动去除位置信息。
      </p>
      {error && <p className="text-sm text-destructive">{error}</p>}
      {value.length > 0 && (
        <Button
          type="button"
          size="xs"
          variant="ghost"
          onClick={() => onChange([])}
        >
          清空图片
        </Button>
      )}
    </div>
  )
}
