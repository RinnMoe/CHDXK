import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useState } from "react"

import { CanteenReviewForm } from "@/components/canteen/canteen-review-form"
import { CanteenMediaUploader } from "@/components/canteen/canteen-media-uploader"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { useCanteenStore } from "@/hooks/use-canteen-catalog"
import { useCreateCanteenReview } from "@/hooks/use-canteen-reviews"
import { canteenDraftKey } from "@/lib/canteen-draft"
import { useAuth } from "@/contexts/auth-context"
import type { CanteenMediaAssetDTO } from "@/api/canteen"

const routeApi = getRouteApi("/app/canteen/stores/$storeID/review/new")

export function CanteenReviewNewPage() {
  const { storeID } = routeApi.useParams()
  const storeIDNumber = Number(storeID)
  const store = useCanteenStore(storeIDNumber)
  const { user } = useAuth()
  const create = useCreateCanteenReview(storeIDNumber)
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)
  const [mediaAssets, setMediaAssets] = useState<CanteenMediaAssetDTO[]>([])

  async function submit(
    values: {
      visitedAt: string
      taste: number
      value: number
      queue: number
      hygiene: number
      service: number
      content: string
      tags: string[]
    },
    idempotencyKey: string
  ) {
    setError(null)
    try {
      await create.mutateAsync({
        input: {
          visited_at: values.visitedAt || undefined,
          taste: values.taste,
          value: values.value,
          queue: values.queue,
          hygiene: values.hygiene,
          service: values.service,
          content: values.content,
          tags: values.tags,
          media_ids: mediaAssets.map((asset) => asset.id),
        },
        idempotencyKey,
      })
      await navigate({
        to: "/canteen/stores/$storeID",
        params: { storeID },
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : "发布评价失败")
    }
  }

  return (
    <>
      <PageTitle>发布到店评价</PageTitle>
      <PageShell>
        <div className="mx-auto max-w-2xl space-y-4">
          <div>
            <h1 className="text-2xl font-bold">发布到店评价</h1>
            <p className="text-sm text-muted-foreground">
              {store.data ? store.data.name : `店铺 #${storeID}`}
            </p>
          </div>
          <CanteenReviewForm
            draftKey={canteenDraftKey(user?.id ?? 0, storeIDNumber, "new")}
            submitLabel="发布评价"
            submitting={create.isPending}
            error={error}
            extra={
              <CanteenMediaUploader
                value={mediaAssets}
                onChange={setMediaAssets}
              />
            }
            onSubmit={(values, key) => void submit(values, key)}
          />
        </div>
      </PageShell>
    </>
  )
}
