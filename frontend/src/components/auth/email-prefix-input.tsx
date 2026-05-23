import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { authEmailDomain, normalizeAuthEmailPrefix } from "@/config/auth"

interface EmailPrefixInputProps {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
  autoComplete?: string
}

export function EmailPrefixInput({
  id,
  label,
  value,
  onChange,
  autoComplete = "username",
}: EmailPrefixInputProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="flex">
        <Input
          id={id}
          type="text"
          inputMode="email"
          placeholder="your"
          value={value}
          onChange={(e) => onChange(normalizeAuthEmailPrefix(e.target.value))}
          autoComplete={autoComplete}
          className="rounded-r-none"
        />
        <span className="inline-flex h-9 shrink-0 items-center rounded-r-md border border-l-0 border-input bg-muted px-3 text-sm text-muted-foreground">
          {authEmailDomain}
        </span>
      </div>
    </div>
  )
}
