import { copy } from "@/lib/copy"
import { Button } from "@/components/ui/button"

export function ReviewSubmitButton({
  submitting,
  isEdit,
}: {
  submitting: boolean
  isEdit: boolean
}) {
  return (
    <Button type="submit" disabled={submitting}>
      {submitting
        ? copy("copy.user.components.review.review-form.2fb1db2875cb")
        : isEdit
          ? copy("copy.user.components.review.review-form.1c8e49eaac67")
          : copy("copy.user.components.review.review-form.ed2f4b0484d7")}
    </Button>
  )
}
