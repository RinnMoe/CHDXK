import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"
import { SafeMarkdown } from "./safe-markdown"
import {
  addTemplateLine,
  CONTENT_MAX_LENGTH,
  CONTENT_MIN_LENGTH,
  hasTemplateLine,
  removeTemplateLine,
  REVIEW_TEMPLATE_LABELS,
} from "./review-form-template"

interface ReviewFormContentEditorProps {
  content: string
  error: string | null
  onChange: (content: string) => void
  onBlur: () => void
}

export function ReviewFormContentEditor({
  content,
  error,
  onChange,
  onBlur,
}: ReviewFormContentEditorProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor="content">
        {copy(
          "copy.user.components.review.review-form-content-editor.0a932e31074d"
        )}
      </Label>
      <div className="flex flex-wrap gap-2">
        {REVIEW_TEMPLATE_LABELS.map((label) => {
          const selected = hasTemplateLine(content, label)
          return (
            <button
              key={label}
              type="button"
              aria-pressed={selected}
              onClick={() => {
                onChange(
                  selected
                    ? removeTemplateLine(content, label)
                    : addTemplateLine(content, label)
                )
              }}
              className={cn(
                "inline-flex h-6 items-center justify-center rounded-4xl border px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-colors focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-none",
                selected
                  ? "border-primary/20 bg-primary/10 text-primary hover:bg-primary/15"
                  : "border-transparent bg-muted text-muted-foreground hover:bg-muted/80 hover:text-foreground"
              )}
            >
              <span>{label.replace(/：$/, "")}</span>
            </button>
          )
        })}
      </div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="space-y-1">
          <p className="text-sm text-muted-foreground">
            {copy(
              "copy.user.components.review.review-form-content-editor.051836569928"
            )}
          </p>
          <Textarea
            id="content"
            placeholder={copy(
              "copy.user.components.review.review-form-content-editor.5774b81023ea"
            )}
            value={content}
            onChange={(e) => onChange(e.target.value)}
            onBlur={onBlur}
            maxLength={CONTENT_MAX_LENGTH}
            rows={10}
            className="resize-y font-sans text-sm"
          />
        </div>
        <div className="space-y-1">
          <p className="text-sm text-muted-foreground">
            {copy(
              "copy.user.components.review.review-form-content-editor.13d61fea9f17"
            )}
          </p>
          <div className="review-markdown prose prose-sm min-h-40 max-w-none dark:prose-invert">
            {content.trim() ? (
              <SafeMarkdown content={content} />
            ) : (
              <span className="text-muted-foreground italic">
                {copy(
                  "copy.user.components.review.review-form-content-editor.d5ee14c78799"
                )}
              </span>
            )}
          </div>
        </div>
      </div>
      <p className="text-sm text-muted-foreground">
        {content.length} / {CONTENT_MAX_LENGTH}
        {copy(
          "copy.user.components.review.review-form-content-editor.ec791ea9ba47"
        )}
        {CONTENT_MIN_LENGTH}
        {copy(
          "copy.user.components.review.review-form-content-editor.457271dfe88b"
        )}
      </p>
      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}
      <div className="text-sm leading-6 text-muted-foreground [&_p]:m-0">
        <p>
          {copy(
            "copy.user.components.review.review-form-content-editor.ac9aab1b8707"
          )}
        </p>
        <p>
          {copy(
            "copy.user.components.review.review-form-content-editor.8031c071e54c"
          )}
        </p>
        <p>
          {copy(
            "copy.user.components.review.review-form-content-editor.6e0358029d6a"
          )}
        </p>
        <p>
          {copy(
            "copy.user.components.review.review-form-content-editor.ab677d30bb6d"
          )}
          <Link to="/faq" className="font-medium text-primary hover:underline">
            {copy(
              "copy.user.components.review.review-form-content-editor.04031807d096"
            )}
          </Link>
          。
        </p>
      </div>
    </div>
  )
}
