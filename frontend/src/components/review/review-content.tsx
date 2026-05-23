import Markdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { Badge } from "@/components/ui/badge"
import { RatingStars } from "./rating-stars"

interface ReviewContentProps {
  rating: number
  semester?: string
  score?: string
  content: string
}

export function ReviewContent({
  rating,
  semester,
  score,
  content,
}: ReviewContentProps) {
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        <RatingStars value={rating} readOnly size="sm" />
        {semester && (
          <Badge variant="outline" className="font-mono">
            {semester}
          </Badge>
        )}
        {score && <Badge variant="secondary">{score}</Badge>}
      </div>

      <div className="prose prose-sm max-w-none text-sm leading-relaxed dark:prose-invert">
        <Markdown remarkPlugins={[remarkGfm]}>{content}</Markdown>
      </div>
    </div>
  )
}
