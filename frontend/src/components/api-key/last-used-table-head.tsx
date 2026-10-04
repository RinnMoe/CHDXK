import { copy } from "@/lib/copy"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

export function LastUsedTableHead() {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            className="inline-flex cursor-help items-center underline decoration-dotted underline-offset-4"
            tabIndex={0}
          >
            {copy(
              "copy.user.components.api-key.last-used-table-head.2b80d4e7163d"
            )}
          </span>
        </TooltipTrigger>
        <TooltipContent>
          {copy(
            "copy.user.components.api-key.last-used-table-head.b7f216012d49"
          )}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
