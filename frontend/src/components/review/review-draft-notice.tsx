import { copy } from "@/lib/copy"
import { Button } from "@/components/ui/button"

export function ReviewDraftNotice({
  visible,
  draftRestored,
  onClear,
}: {
  visible: boolean
  draftRestored: boolean
  onClear: () => void
}) {
  return (
    <>
      {visible && (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
          <span>
            {draftRestored
              ? copy("copy.user.components.review.review-form.bfacfa460b1c")
              : copy("copy.user.components.review.review-form.e9643cc29816")}
          </span>
          <Button type="button" variant="ghost" size="sm" onClick={onClear}>
            {copy("copy.user.components.review.review-form.df8ab5bdc788")}
          </Button>
        </div>
      )}
    </>
  )
}
