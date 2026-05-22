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
import { Textarea } from "@/components/ui/textarea"
import { RatingStars } from "./rating-stars"
import type {
  CreateReviewCommand,
  UpdateReviewCommand,
  ReviewDTO,
} from "@/api/review"

interface ReviewFormProps {
  courseID?: number
  initialReview?: ReviewDTO
  semesters?: string[]
  onSubmit: (cmd: CreateReviewCommand | UpdateReviewCommand) => Promise<void> | void
  onCancel?: () => void
  isSubmitting?: boolean
}

export function ReviewForm({
  courseID,
  initialReview,
  semesters,
  onSubmit,
  onCancel,
  isSubmitting,
}: ReviewFormProps) {
  const isEdit = !!initialReview
  const [rating, setRating] = useState(initialReview?.rating ?? 0)
  const [content, setContent] = useState(initialReview?.content ?? "")
  const [semester, setSemester] = useState(initialReview?.semester ?? "")
  const [score, setScore] = useState(initialReview?.score ?? "")
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!semester) {
      setError("请选择学期")
      return
    }
    if (rating < 1 || rating > 5) {
      setError("请选择评分（1-5 星）")
      return
    }
    if (content.trim().length < 10) {
      setError("评价内容至少需要 10 个字符")
      return
    }
    try {
      if (isEdit) {
        await onSubmit({
          semester,
          rating,
          content,
          score: score || undefined,
        } as UpdateReviewCommand)
      } else {
        if (!courseID) throw new Error("missing courseID")
        await onSubmit({
          course_id: courseID,
          semester,
          rating,
          content,
          score: score || undefined,
        } as CreateReviewCommand)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "提交失败")
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <div className="space-y-2">
        <Label>评分</Label>
        <div>
          <RatingStars value={rating} onChange={setRating} size="lg" />
        </div>
      </div>

      <div className="grid sm:grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>学期</Label>
          <Select value={semester} onValueChange={setSemester}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="选择学期" />
            </SelectTrigger>
            <SelectContent>
              {semesters?.map((s) => (
                <SelectItem key={s} value={s}>
                  {s}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="score">分数（可选）</Label>
          <Input
            id="score"
            placeholder="如 A、92、未公布"
            value={score}
            onChange={(e) => setScore(e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="content">评价内容</Label>
        <Textarea
          id="content"
          placeholder="分享你对这门课程的看法..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={6}
          className="resize-y"
        />
        <p className="text-xs text-muted-foreground">{content.length} / 至少 10 字</p>
      </div>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      <div className="flex gap-2 justify-end">
        {onCancel && (
          <Button type="button" variant="ghost" onClick={onCancel}>
            取消
          </Button>
        )}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : isEdit ? "更新评价" : "发布评价"}
        </Button>
      </div>
    </form>
  )
}
