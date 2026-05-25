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
            最后使用
          </span>
        </TooltipTrigger>
        <TooltipContent>数据延迟约5分钟</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
