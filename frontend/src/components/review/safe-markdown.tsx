import { Link } from "@tanstack/react-router"
import Markdown from "react-markdown"
import rehypeSanitize from "rehype-sanitize"
import remarkBreaks from "remark-breaks"
import remarkGfm from "remark-gfm"
import type { Components } from "react-markdown"

interface MarkdownNode {
  type: string
  value?: string
  url?: string
  title?: string | null
  children?: MarkdownNode[]
}

interface SafeMarkdownProps {
  content: string
}

const reviewReferencePattern = /(^|[^A-Za-z0-9_/#])#([1-9]\d*)\b/g

function replaceReviewReferences(value: string) {
  const nodes: MarkdownNode[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null

  reviewReferencePattern.lastIndex = 0

  while ((match = reviewReferencePattern.exec(value)) !== null) {
    const [fullMatch, prefix, reviewID] = match
    const referenceStart = match.index + prefix.length

    if (referenceStart > lastIndex) {
      nodes.push({
        type: "text",
        value: value.slice(lastIndex, referenceStart),
      })
    }

    nodes.push({
      type: "link",
      url: `/review/${reviewID}`,
      title: null,
      children: [{ type: "text", value: `#${reviewID}` }],
    })

    lastIndex = match.index + fullMatch.length
  }

  if (nodes.length === 0) return null

  if (lastIndex < value.length) {
    nodes.push({ type: "text", value: value.slice(lastIndex) })
  }

  return nodes
}

function transformReviewReferences(node: MarkdownNode) {
  if (!node.children || node.type === "link" || node.type === "linkReference") {
    return
  }

  const children: MarkdownNode[] = []
  let changed = false

  for (const child of node.children) {
    if (child.type === "text" && child.value) {
      const replacement = replaceReviewReferences(child.value)
      if (replacement) {
        children.push(...replacement)
        changed = true
        continue
      }
    }

    transformReviewReferences(child)
    children.push(child)
  }

  if (changed) node.children = children
}

function remarkReviewReferences() {
  return function transform(tree: MarkdownNode) {
    transformReviewReferences(tree)
  }
}

const markdownComponents: Components = {
  hr() {
    return <hr />
  },
  a({ href, children, title }) {
    if (href?.startsWith("/review/")) {
      return (
        <Link to={href} title={title}>
          {children}
        </Link>
      )
    }

    return (
      <a href={href} title={title}>
        {children}
      </a>
    )
  },
}

export function SafeMarkdown({ content }: SafeMarkdownProps) {
  return (
    <Markdown
      components={markdownComponents}
      remarkPlugins={[remarkGfm, remarkBreaks, remarkReviewReferences]}
      rehypePlugins={[rehypeSanitize]}
      skipHtml
    >
      {content}
    </Markdown>
  )
}
