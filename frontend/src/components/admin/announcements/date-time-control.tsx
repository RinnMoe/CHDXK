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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { formatDateTime } from "@/lib/date"

const hourOptions = Array.from({ length: 24 }, (_, i) =>
  String(i).padStart(2, "0")
)
const minuteOptions = Array.from({ length: 60 }, (_, i) =>
  String(i).padStart(2, "0")
)

interface DateTimeControlProps {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
}

export function DateTimeControl({
  id,
  label,
  value,
  onChange,
}: DateTimeControlProps) {
  const [open, setOpen] = useState(false)
  const current = dayjs(value)
  const selected = current.isValid() ? current.toDate() : undefined
  const hour = current.isValid() ? current.format("HH") : "00"
  const minute = current.isValid() ? current.format("mm") : "00"

  function nextValue(date: Date, nextHour = hour, nextMinute = minute) {
    return dayjs(date)
      .hour(Number(nextHour))
      .minute(Number(nextMinute))
      .second(0)
      .millisecond(0)
      .format("YYYY-MM-DDTHH:mm")
  }

  function setTime(nextHour: string, nextMinute: string) {
    onChange(nextValue(selected ?? new Date(), nextHour, nextMinute))
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={id}
            type="button"
            variant="outline"
            className="h-9 w-full justify-between font-normal"
          >
            <span className={selected ? undefined : "text-muted-foreground"}>
              {selected ? formatDateTime(selected) : "选择时间"}
            </span>
            <RiCalendarLine className="size-4 text-muted-foreground" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={selected}
            defaultMonth={selected}
            onSelect={(date) => {
              if (!date) return
              onChange(nextValue(date))
            }}
            locale={zhCN}
            captionLayout="dropdown"
          />
          <div className="flex items-center gap-2 border-t p-3">
            <Select
              value={hour}
              onValueChange={(nextHour) => setTime(nextHour, minute)}
            >
              <SelectTrigger className="w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {hourOptions.map((item) => (
                  <SelectItem key={item} value={item}>
                    {item} 时
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={minute}
              onValueChange={(nextMinute) => setTime(hour, nextMinute)}
            >
              <SelectTrigger className="w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {minuteOptions.map((item) => (
                  <SelectItem key={item} value={item}>
                    {item} 分
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setOpen(false)}
            >
              完成
            </Button>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
