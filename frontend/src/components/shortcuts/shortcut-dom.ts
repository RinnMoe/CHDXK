import {
  ACCESS_KEYS,
  SHORTCUT_ROOT_SELECTOR,
  type AccessHint,
} from "./shortcut-data"

const ACCESS_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "input:not([type='hidden']):not([disabled])",
  "textarea:not([disabled])",
  "select:not([disabled])",
  "[role='button']:not([aria-disabled='true'])",
  "[role='menuitem']:not([aria-disabled='true'])",
].join(",")

export function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) {
    return false
  }

  if (target.isContentEditable) {
    return true
  }

  return Boolean(target.closest("input, textarea, select, [contenteditable]"))
}

function isVisibleElement(element: HTMLElement) {
  if (element.closest(SHORTCUT_ROOT_SELECTOR)) {
    return false
  }

  if (element.getAttribute("aria-hidden") === "true") {
    return false
  }

  const style = window.getComputedStyle(element)
  if (style.visibility === "hidden" || style.display === "none") {
    return false
  }

  const rect = element.getBoundingClientRect()
  return (
    rect.width > 0 &&
    rect.height > 0 &&
    rect.bottom >= 0 &&
    rect.right >= 0 &&
    rect.top <= window.innerHeight &&
    rect.left <= window.innerWidth
  )
}

function getElementLabel(element: HTMLElement) {
  const ariaLabel = element.getAttribute("aria-label")?.trim()
  if (ariaLabel) return ariaLabel

  const title = element.getAttribute("title")?.trim()
  if (title) return title

  const placeholder = element.getAttribute("placeholder")?.trim()
  if (placeholder) return placeholder

  const text = element.textContent?.replace(/\s+/g, " ").trim()
  if (text) return text.slice(0, 32)

  if (element instanceof HTMLInputElement) {
    return element.name || "输入框"
  }

  return "控件"
}

export function collectAccessHints() {
  const candidates = Array.from(
    document.querySelectorAll<HTMLElement>(ACCESS_SELECTOR)
  )

  return candidates
    .filter(isVisibleElement)
    .slice(0, ACCESS_KEYS.length)
    .map((element, index) => ({
      key: ACCESS_KEYS[index] ?? "",
      label: getElementLabel(element),
      element,
      rect: element.getBoundingClientRect(),
    }))
}

function isFocusableControl(element: HTMLElement) {
  return (
    element instanceof HTMLInputElement ||
    element instanceof HTMLTextAreaElement ||
    element instanceof HTMLSelectElement
  )
}

export function activateHint(hint: AccessHint) {
  if (isFocusableControl(hint.element)) {
    hint.element.focus()
    if (
      hint.element instanceof HTMLInputElement ||
      hint.element instanceof HTMLTextAreaElement
    ) {
      hint.element.select()
    }
    return
  }

  hint.element.click()
}
