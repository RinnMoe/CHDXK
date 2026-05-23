import { useMemo, useState } from "react"
import { Link } from "react-router-dom"
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
import { SafeMarkdown } from "./safe-markdown"
import type {
  CreateReviewCommand,
  UpdateReviewCommand,
  ReviewDTO,
} from "@/api/review"

interface ReviewFormProps {
  courseID?: number
  initialReview?: ReviewDTO
  semesters?: string[]
  onSubmit: (
    cmd: CreateReviewCommand | UpdateReviewCommand
  ) => Promise<void> | void
  onCancel?: () => void
  isSubmitting?: boolean
}

const SCORE_MAX_LENGTH = 10
const CONTENT_MIN_LENGTH = 4
const CONTENT_MAX_LENGTH = 9691
const DEFAULT_REVIEW_TEMPLATE = `课程内容：

上课自由度：

考核标准：

授课质量：`

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
  const [content, setContent] = useState(
    initialReview?.content ?? DEFAULT_REVIEW_TEMPLATE
  )
  const [semester, setSemester] = useState(initialReview?.semester ?? "")
  const [score, setScore] = useState(initialReview?.score ?? "")
  const [error, setError] = useState<string | null>(null)
  const availableSemesters = useMemo(() => {
    const values = semesters ? [...semesters] : []
    const initialSemester = initialReview?.semester
    if (initialSemester && !values.includes(initialSemester)) {
      values.unshift(initialSemester)
    }
    return values
  }, [initialReview?.semester, semesters])

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
    if (score.length > SCORE_MAX_LENGTH) {
      setError(`分数最多 ${SCORE_MAX_LENGTH} 个字符`)
      return
    }
    if (!isEdit && content.trim() === DEFAULT_REVIEW_TEMPLATE.trim()) {
      setError("请修改点评模板后再提交")
      return
    }
    if (content.trim().length < CONTENT_MIN_LENGTH) {
      setError(`点评内容至少需要 ${CONTENT_MIN_LENGTH} 个字符`)
      return
    }
    if (content.length > CONTENT_MAX_LENGTH) {
      setError(`点评内容最多 ${CONTENT_MAX_LENGTH} 个字符`)
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

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>学期</Label>
          <Select value={semester} onValueChange={setSemester}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="选择学期" />
            </SelectTrigger>
            <SelectContent>
              {availableSemesters.map((s) => (
                <SelectItem key={s} value={s}>
                  {s}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-sm leading-6 text-muted-foreground">
            2026-2027 代表 2026-2027 学年度（2026.9-2027.8）。1代表秋季学期，2代表春季学期，3代表夏季学期/小学期。
          </p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="score">分数（可选）</Label>
          <Input
            id="score"
            placeholder="如 A、92、未公布"
            value={score}
            onChange={(e) => setScore(e.target.value)}
            maxLength={SCORE_MAX_LENGTH}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="content">点评内容</Label>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <p className="text-sm text-muted-foreground">编辑</p>
            <Textarea
              id="content"
              placeholder="分享你对这门课程的看法...（支持 Markdown）"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              maxLength={CONTENT_MAX_LENGTH}
              rows={10}
              className="resize-y font-mono text-sm"
            />
          </div>
          <div className="space-y-1">
            <p className="text-sm text-muted-foreground">预览</p>
            <div className="prose prose-sm min-h-[10rem] max-w-none text-sm dark:prose-invert">
              {content.trim() ? (
                <SafeMarkdown content={content} />
              ) : (
                <span className="text-muted-foreground italic">预览区域</span>
              )}
            </div>
          </div>
        </div>
        <p className="text-sm text-muted-foreground">
          {content.length} / {CONTENT_MAX_LENGTH} 字，至少 {CONTENT_MIN_LENGTH} 字
        </p>
        <div className="text-sm leading-6 text-muted-foreground [&_p]:m-0">
          <p>
            欢迎畅所欲言。点评模板可以按需修改或删除。编辑框支持 Markdown
            语法。
          </p>
          <p>
            理想的点评应当富有事实且对课程有全面的描述。比如课讲得好但是考核很严格，或者作业奇葩但给分很高。二者都说出来更有利于同学们做出全面的选择和判断。
          </p>
          <p>
            避免滥用缩写、梗、隐喻等让其他读者难以理解的表达方式和内容。避免使用情绪化用语和冒犯性言论。
          </p>
          <p>
            提交点评表示您同意授权本网站使用点评的内容，并且了解本站的
            <Link to="/faq" className="font-medium text-primary hover:underline">
              相关立场
            </Link>
            。
          </p>
        </div>
      </div>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      <div className="flex justify-end gap-2">
        {onCancel && (
          <Button type="button" variant="ghost" onClick={onCancel}>
            取消
          </Button>
        )}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : isEdit ? "更新点评" : "发布点评"}
        </Button>
      </div>
    </form>
  )
}
