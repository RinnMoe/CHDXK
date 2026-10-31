import { Button } from "@/components/ui/button"

export function CanteenRatingInput({
  label,
  value,
  onChange,
}: {
  label: string
  value: number
  onChange: (value: number) => void
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-sm text-muted-foreground">{label}</span>
      <div className="flex gap-1">
        {[1, 2, 3, 4, 5].map((score) => (
          <Button
            key={score}
            type="button"
            size="xs"
            variant={score === value ? "default" : "outline"}
            aria-label={`${label} ${score} 分`}
            aria-pressed={score === value}
            onClick={() => onChange(score)}
          >
            {score}
          </Button>
        ))}
      </div>
    </div>
  )
}
