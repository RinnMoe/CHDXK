export const SCORE_MAX_LENGTH = 10
export const CONTENT_MIN_LENGTH = 4
export const CONTENT_MAX_LENGTH = 9681

export const REVIEW_TEMPLATE_LABELS: readonly string[] = [
  "课程内容：",
  "上课自由度：",
  "考核标准：",
  "授课质量：",
] as const

export const DEFAULT_REVIEW_TEMPLATE = REVIEW_TEMPLATE_LABELS.join("\n\n")

const TEXT_CONTENT_PATTERN = /[\p{L}\p{N}]/u

function findTemplateLineIndex(lines: string[], label: string) {
  return lines.findIndex((line) => line.trimStart().startsWith(label))
}

export function hasTemplateLine(content: string, label: string) {
  return findTemplateLineIndex(content.split("\n"), label) >= 0
}

function templateLineHasUserInput(line: string, label: string) {
  const labelIndex = line.indexOf(label)
  if (labelIndex < 0) return false
  return line.slice(labelIndex + label.length).trim().length > 0
}

function removeTemplateLabel(line: string) {
  const label = REVIEW_TEMPLATE_LABELS.find((templateLabel) =>
    line.trimStart().startsWith(templateLabel)
  )

  return label ? line.trimStart().slice(label.length) : line
}

export function hasActualReviewContent(content: string) {
  return content
    .split("\n")
    .some((line) => TEXT_CONTENT_PATTERN.test(removeTemplateLabel(line)))
}

export function addTemplateLine(content: string, label: string) {
  if (hasTemplateLine(content, label)) return content

  if (content.trim().length === 0) {
    return label
  }

  const lines = content.replace(/\s*$/, "").split("\n")
  const newLabelOrder = REVIEW_TEMPLATE_LABELS.indexOf(label)
  const nextTemplateLineIndex = lines.findIndex((line) => {
    const templateOrder = REVIEW_TEMPLATE_LABELS.findIndex((templateLabel) =>
      line.trimStart().startsWith(templateLabel)
    )
    return templateOrder > newLabelOrder
  })

  if (nextTemplateLineIndex < 0) {
    return `${lines.join("\n")}\n\n${label}`
  }

  lines.splice(nextTemplateLineIndex, 0, label, "")
  return lines.join("\n")
}

export function removeTemplateLine(content: string, label: string) {
  const lines = content.split("\n")
  const lineIndex = findTemplateLineIndex(lines, label)
  if (lineIndex < 0) return content
  if (templateLineHasUserInput(lines[lineIndex], label)) return content

  lines.splice(lineIndex, 1)

  while (
    lineIndex < lines.length &&
    lines[lineIndex] === "" &&
    (lineIndex === 0 || lines[lineIndex - 1] === "")
  ) {
    lines.splice(lineIndex, 1)
  }

  while (
    lineIndex > 0 &&
    lineIndex === lines.length &&
    lines[lineIndex - 1] === "" &&
    (lineIndex - 1 === 0 || lines[lineIndex - 2] === "")
  ) {
    lines.splice(lineIndex - 1, 1)
  }

  return lines.join("\n")
}
