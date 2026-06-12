import { createPortal } from "react-dom"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Kbd, KbdGroup } from "@/components/ui/kbd"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import {
  actionShortcuts,
  pageShortcuts,
  type AccessHint,
  type PageShortcut,
} from "./shortcut-data"

function ShortcutKeys({ keys }: { keys: string }) {
  return (
    <KbdGroup className="shrink-0">
      {keys.split(" ").map((key) => (
        <Kbd key={key} className="min-w-6 text-foreground">
          {key}
        </Kbd>
      ))}
    </KbdGroup>
  )
}

function ShortcutRow({
  keys,
  title,
  description,
  disabled,
}: {
  keys: string
  title: string
  description: string
  disabled?: boolean
}) {
  return (
    <div
      className={cn(
        "grid grid-cols-[4.75rem_minmax(0,1fr)] items-start gap-3 rounded-md p-2",
        disabled && "opacity-45"
      )}
    >
      <ShortcutKeys keys={keys} />
      <div className="min-w-0">
        <div className="font-medium text-foreground">{title}</div>
        <div className="text-xs leading-5 text-muted-foreground">
          {description}
        </div>
      </div>
    </div>
  )
}

export function AccessHintOverlay({ hints }: { hints: AccessHint[] }) {
  return createPortal(
    <div
      data-shortcut-root
      className="pointer-events-none fixed inset-0 z-70"
      aria-hidden="true"
    >
      <div className="absolute top-16 left-1/2 -translate-x-1/2 rounded-md border bg-popover px-3 py-2 text-xs text-popover-foreground shadow-lg">
        输入标记可点击或聚焦控件，Esc 退出
      </div>
      {hints.map((hint) => (
        <Kbd
          key={`${hint.key}-${hint.label}-${hint.rect.left}-${hint.rect.top}`}
          className="absolute border border-primary bg-primary px-1.5 font-mono font-bold text-primary-foreground shadow-lg"
          style={{
            top: Math.max(4, hint.rect.top - 8),
            left: Math.max(4, hint.rect.left - 8),
          }}
        >
          {hint.key.toUpperCase()}
        </Kbd>
      ))}
    </div>,
    document.body
  )
}

export function ShortcutHelpDialog({
  open,
  onOpenChange,
  availablePageShortcuts,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  availablePageShortcuts: PageShortcut[]
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        data-shortcut-root
        className="max-h-[min(42rem,calc(100svh-2rem))] grid-rows-[auto_minmax(0,1fr)] overflow-hidden sm:max-w-2xl"
      >
        <DialogHeader>
          <DialogTitle>快捷键</DialogTitle>
          <DialogDescription>
            页面跳转可先按 G 再按目标键；输入框内不会触发全局快捷键。
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="-mx-2 min-h-0 pr-3">
          <div className="space-y-5 px-2 pb-1">
            <section>
              <h3 className="mb-2 text-sm font-medium">页面</h3>
              <div className="space-y-1">
                {pageShortcuts.map((shortcut) => {
                  const disabled = !availablePageShortcuts.includes(shortcut)
                  return (
                    <ShortcutRow
                      key={shortcut.path}
                      keys={shortcut.keys}
                      title={shortcut.title}
                      description={
                        disabled
                          ? shortcut.requiresAdmin
                            ? "需要管理员权限"
                            : "需要登录"
                          : shortcut.description
                      }
                      disabled={disabled}
                    />
                  )
                })}
              </div>
            </section>

            <section>
              <h3 className="mb-2 text-sm font-medium">操作</h3>
              <div className="space-y-1">
                {actionShortcuts.map((shortcut) => (
                  <ShortcutRow
                    key={shortcut.keys}
                    keys={shortcut.keys}
                    title={shortcut.title}
                    description={shortcut.description}
                  />
                ))}
              </div>
            </section>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
