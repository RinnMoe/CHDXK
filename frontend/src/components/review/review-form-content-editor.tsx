import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { Label } from "@/components/ui/label"
import MDEditor, {
  commands,
  type ICommand,
} from "@uiw/react-md-editor/nohighlight"
import "@uiw/react-md-editor/markdown-editor.css"
import "./review-editor.css"
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

function localize(command: ICommand, label: string): ICommand {
  return {
    ...command,
    buttonProps: { ...command.buttonProps, title: label, "aria-label": label },
  }
}
const editorCommands = [
  localize(commands.bold, "加粗"),
  localize(commands.italic, "斜体"),
  localize(commands.link, "插入链接"),
  localize(commands.quote, "引用"),
  localize(commands.unorderedListCommand, "无序列表"),
  localize(commands.orderedListCommand, "有序列表"),
]
const previewCommands = [
  { ...localize(commands.codeEdit, "编辑"), icon: <span>编辑</span> },
  { ...localize(commands.codePreview, "预览"), icon: <span>预览</span> },
]

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
      <MDEditor
        data-review-editor=""
        value={content}
        onChange={(value) => onChange(value ?? "")}
        height={260}
        preview="edit"
        visibleDragbar={false}
        highlightEnable={false}
        commands={editorCommands}
        extraCommands={previewCommands}
        textareaProps={{
          id: "content",
          "aria-label": "点评内容",
          "aria-invalid": !!error,
          maxLength: CONTENT_MAX_LENGTH,
          onBlur,
          placeholder: "分享课程内容、考核方式或你的上课体验…",
        }}
        components={{
          preview: (source) => (
            <div className="review-markdown prose prose-sm max-w-none wrap-break-word dark:prose-invert">
              <SafeMarkdown content={source} />
            </div>
          ),
        }}
      />
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
      <details className="text-xs leading-6 text-muted-foreground [&_p]:m-0">
        <summary className="cursor-pointer font-medium">
          点评建议与发布须知
        </summary>
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
      </details>
    </div>
  )
}
