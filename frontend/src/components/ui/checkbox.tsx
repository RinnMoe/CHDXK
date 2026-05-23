import * as React from "react"
import { Checkbox as CheckboxPrimitive } from "radix-ui"

import { cn } from "@/lib/utils"
import { RiCheckLine } from "@remixicon/react"

function Checkbox({
  className,
  ...props
}: React.ComponentProps<typeof CheckboxPrimitive.Root>) {
  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      className={cn(
        "peer data-checked:animate-checkbox-pop relative flex size-4 shrink-0 items-center justify-center rounded-[4px] border border-input shadow-xs transition-[color,background-color,border-color,box-shadow,transform] duration-150 ease-out outline-none group-has-disabled/field:opacity-50 after:absolute after:-inset-x-3 after:-inset-y-2 hover:border-ring/70 hover:bg-muted/30 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:scale-90 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:border-input disabled:hover:bg-transparent aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 aria-invalid:aria-checked:border-primary dark:bg-input/30 dark:hover:bg-input/40 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 data-checked:border-primary data-checked:bg-primary data-checked:text-primary-foreground data-checked:hover:bg-primary/90 data-checked:hover:text-primary-foreground dark:data-checked:bg-primary dark:data-checked:hover:bg-primary/90",
        className
      )}
      {...props}
    >
      <CheckboxPrimitive.Indicator
        data-slot="checkbox-indicator"
        className="grid animate-in place-content-center text-current duration-150 zoom-in-50 [&>svg]:size-3.5"
      >
        <RiCheckLine />
      </CheckboxPrimitive.Indicator>
    </CheckboxPrimitive.Root>
  )
}

export { Checkbox }
