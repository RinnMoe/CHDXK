import { copy } from "@/lib/copy"
import { useState } from "react"
import dayjs from "dayjs"
import { zhCN } from "date-fns/locale"
import { RiCalendarLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { formatDateInputValue } from "@/lib/date"

function parseDateInputValue(value: string): Date | undefined {
  const date = dayjs(value)
  return date.isValid() ? date.toDate() : undefined
}

interface DateFilterPickerProps {
  id: string
  name: string
  label: string
  value: string
  onChange: (value: string) => void
}

export function DateFilterPicker({
  id,
  name,
  label,
  value,
  onChange,
}: DateFilterPickerProps) {
  const [open, setOpen] = useState(false)
  const selectedDate = parseDateInputValue(value)

  return (
    <div className="flex w-40 flex-col gap-1">
      <Label htmlFor={id}>{label}</Label>
      <input type="hidden" name={name} value={value} />
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={id}
            type="button"
            variant="outline"
            className="h-9 w-full justify-between font-normal"
          >
            <span className={value ? undefined : "text-muted-foreground"}>
              {value ||
                copy(
                  "copy.admin.components.admin.audit-logs.date-filter-picker.680c78ec60f4"
                )}
            </span>
            <RiCalendarLine className="size-4 text-muted-foreground" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={selectedDate}
            defaultMonth={selectedDate}
            onSelect={(date) => {
              if (!date) return
              onChange(formatDateInputValue(date))
              setOpen(false)
            }}
            locale={zhCN}
            captionLayout="dropdown"
          />
          {value ? (
            <div className="border-t p-2">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="w-full"
                onClick={() => {
                  onChange("")
                  setOpen(false)
                }}
              >
                {copy(
                  "copy.admin.components.admin.audit-logs.date-filter-picker.d68c450b1104"
                )}
              </Button>
            </div>
          ) : null}
        </PopoverContent>
      </Popover>
    </div>
  )
}
