import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useState } from "react"

import { CanteenReviewForm } from "@/components/canteen/canteen-review-form"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { useAuth } from "@/contexts/auth-context"
import { useMyCanteenReviews, useUpdateCanteenReview } from "@/hooks/use-canteen-reviews"
import { canteenDraftKey } from "@/lib/canteen-draft"

const routeApi = getRouteApi("/app/canteen/reviews/$reviewID/edit")

export function CanteenReviewEditPage() {
  const { reviewID } = routeApi.useParams()
  const reviewIDNumber = Number(reviewID)
  const mine = useMyCanteenReviews(1, 100)
  const update = useUpdateCanteenReview(reviewIDNumber)
  const { user } = useAuth()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  const review = mine.data?.items.find((item) => item.id === reviewIDNumber)

  if (mine.isLoading) {
    return (
      <PageShell>
        <p className="py-16 text-center text-sm text-muted-foreground">
          正在加载评价…
        </p>
      </PageShell>
    )
  }

  if (!review) {
    return (
      <PageShell>
        <div className="space-y-3 py-16 text-center">
          <p className="text-sm font-medium">评价不存在或不可编辑</p>
          <button
            type="button"
            className="text-sm text-primary hover:underline"
            onClick={() => void navigate({ to: "/canteen/reviews/mine" })}
          >
            返回我的评价
          </button>
        </div>
      </PageShell>
    )
  }

  const currentReview = review

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
    _idempotencyKey: string
  ) {
    setError(null)
    try {
      await update.mutateAsync({
        visited_at: values.visitedAt || undefined,
        taste: values.taste,
        value: values.value,
        queue: values.queue,
        hygiene: values.hygiene,
        service: values.service,
        content: values.content,
        tags: values.tags,
        expected_updated_at: currentReview.updated_at,
      })
      await navigate({ to: "/canteen/reviews/mine" })
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存评价失败")
    }
  }

  return (
    <>
      <PageTitle>修改到店评价</PageTitle>
      <PageShell>
        <div className="mx-auto max-w-2xl space-y-4">
          <div>
            <h1 className="text-2xl font-bold">修改到店评价</h1>
            <p className="text-sm text-muted-foreground">
              {currentReview.store_name}
            </p>
          </div>
          <CanteenReviewForm
            draftKey={canteenDraftKey(
              user?.id ?? 0,
              currentReview.store_id,
              `edit:${currentReview.id}`
            )}
            initial={{
              visitedAt: currentReview.visited_at,
              taste: currentReview.taste,
              value: currentReview.value,
              queue: currentReview.queue,
              hygiene: currentReview.hygiene,
              service: currentReview.service,
              content: currentReview.content,
              tags: currentReview.tags,
            }}
            submitLabel="保存修改"
            submitting={update.isPending}
            error={error}
            onSubmit={(values, key) => void submit(values, key)}
          />
        </div>
      </PageShell>
    </>
  )
}
