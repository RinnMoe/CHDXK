import { useEffect, useState } from "react"
import { useDebounce } from "use-debounce"
import { Input } from "@/components/ui/input"

type DebouncedSearchInputProps = {
  value: string
  placeholder?: string
  debounceMs: number
  onDebouncedChange: (value: string) => void
}

export function DebouncedSearchInput({
  value,
  placeholder,
  debounceMs,
  onDebouncedChange,
}: DebouncedSearchInputProps) {
  const [inputValue, setInputValue] = useState(value)
  const [lastValue, setLastValue] = useState(value)
  const [isComposing, setIsComposing] = useState(false)
  const [debouncedValue] = useDebounce(inputValue, debounceMs)

  if (!isComposing && value !== lastValue) {
    setLastValue(value)
    setInputValue(value)
  }

  useEffect(() => {
    if (isComposing || debouncedValue !== inputValue) return
    onDebouncedChange(debouncedValue)
  }, [debouncedValue, inputValue, isComposing, onDebouncedChange])

  return (
    <Input
      placeholder={placeholder}
      value={inputValue}
      onChange={(e) => setInputValue(e.target.value)}
      onCompositionStart={() => setIsComposing(true)}
      onCompositionEnd={(e) => {
        setIsComposing(false)
        setInputValue(e.currentTarget.value)
      }}
    />
  )
}
