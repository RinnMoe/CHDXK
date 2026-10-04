import { copy } from "@/lib/copy"
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
        {copy(
          "copy.user.components.shortcuts.shortcut-help-dialog.ff5a83daebf9"
        )}
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
          <DialogTitle>
            {copy(
              "copy.user.components.shortcuts.shortcut-help-dialog.ee2638183d3e"
            )}
          </DialogTitle>
          <DialogDescription>
            {copy(
              "copy.user.components.shortcuts.shortcut-help-dialog.93c6900d289b"
            )}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="-mx-2 min-h-0 pr-3">
          <div className="space-y-5 px-2 pb-1">
            <section>
              <h3 className="mb-2 text-sm font-medium">
                {copy(
                  "copy.user.components.shortcuts.shortcut-help-dialog.452c7b10d57a"
                )}
              </h3>
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
                            ? copy(
                                "copy.user.components.shortcuts.shortcut-help-dialog.9228f9aa4c65"
                              )
                            : copy(
                                "copy.user.components.shortcuts.shortcut-help-dialog.6eb1b64e260a"
                              )
                          : shortcut.description
                      }
                      disabled={disabled}
                    />
                  )
                })}
              </div>
            </section>

            <section>
              <h3 className="mb-2 text-sm font-medium">
                {copy(
                  "copy.user.components.shortcuts.shortcut-help-dialog.ed31fbb483ee"
                )}
              </h3>
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
