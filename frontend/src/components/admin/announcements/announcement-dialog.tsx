import { copy } from "@/lib/copy"
import { useState, type SyntheticEvent } from "react"
import { useForm } from "@tanstack/react-form"
import type {
  AnnouncementDTO,
  SaveAnnouncementCommand,
} from "@/api/announcement"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { getErrorMessage } from "@/components/admin/admin-utils"
import {
  buildAnnouncementCommand,
  toAnnouncementFormState,
  validateAnnouncementForm,
} from "./announcement-form"
import { DateTimeControl } from "./date-time-control"

interface AnnouncementDialogProps {
  open: boolean
  announcement: AnnouncementDTO | null
  isSaving: boolean
  onOpenChange: (open: boolean) => void
  onSave: (cmd: SaveAnnouncementCommand) => Promise<void>
}

export function AnnouncementDialog({
  open,
  announcement,
  isSaving,
  onOpenChange,
  onSave,
}: AnnouncementDialogProps) {
  const [submitError, setSubmitError] = useState("")
  const form = useForm({
    defaultValues: toAnnouncementFormState(announcement),
    onSubmit: async ({ value }) => {
      const validationError = validateAnnouncementForm(value)
      if (validationError) throw new Error(validationError)

      await onSave(buildAnnouncementCommand(value))
      onOpenChange(false)
    },
  })

  function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    event.stopPropagation()
    setSubmitError("")
    void form.handleSubmit().catch((err: unknown) => {
      setSubmitError(getErrorMessage(err))
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {announcement
              ? copy(
                  "copy.admin.components.admin.announcements.announcement-dialog.83163278a659"
                )
              : copy(
                  "copy.admin.components.admin.announcements.announcement-dialog.ee4ea097b207"
                )}
          </DialogTitle>
        </DialogHeader>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="grid gap-4 sm:grid-cols-2">
            <form.Field name="title">
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor="announcement-title">
                    {copy(
                      "copy.admin.components.admin.announcements.announcement-dialog.34638cbed99a"
                    )}
                  </Label>
                  <Input
                    id="announcement-title"
                    value={field.state.value}
                    maxLength={200}
                    onChange={(event) => field.handleChange(event.target.value)}
                    onBlur={field.handleBlur}
                  />
                </div>
              )}
            </form.Field>
            <form.Field name="priority">
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor="announcement-priority">
                    {copy(
                      "copy.admin.components.admin.announcements.announcement-dialog.bb8e192e1eed"
                    )}
                  </Label>
                  <Input
                    id="announcement-priority"
                    type="number"
                    value={field.state.value}
                    onChange={(event) => field.handleChange(event.target.value)}
                    onBlur={field.handleBlur}
                  />
                </div>
              )}
            </form.Field>
            <form.Field name="body">
              {(field) => (
                <div className="space-y-2 sm:col-span-2">
                  <Label htmlFor="announcement-body">
                    {copy(
                      "copy.admin.components.admin.announcements.announcement-dialog.76c467671ab1"
                    )}
                  </Label>
                  <Textarea
                    id="announcement-body"
                    value={field.state.value}
                    className="min-h-28"
                    onChange={(event) => field.handleChange(event.target.value)}
                    onBlur={field.handleBlur}
                  />
                </div>
              )}
            </form.Field>
            <form.Field name="link_title">
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor="announcement-link-title">
                    {copy(
                      "copy.admin.components.admin.announcements.announcement-dialog.db21da93d322"
                    )}
                  </Label>
                  <Input
                    id="announcement-link-title"
                    value={field.state.value}
                    maxLength={200}
                    onChange={(event) => field.handleChange(event.target.value)}
                    onBlur={field.handleBlur}
                  />
                </div>
              )}
            </form.Field>
            <form.Field name="link_url">
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor="announcement-link-url">
                    {copy(
                      "copy.admin.components.admin.announcements.announcement-dialog.a48ce0b1dc35"
                    )}
                  </Label>
                  <Input
                    id="announcement-link-url"
                    value={field.state.value}
                    placeholder={copy(
                      "copy.admin.components.admin.announcements.announcement-dialog.100680ad546c"
                    )}
                    onChange={(event) => field.handleChange(event.target.value)}
                    onBlur={field.handleBlur}
                  />
                </div>
              )}
            </form.Field>
            <form.Field name="show_start">
              {(field) => (
                <DateTimeControl
                  id="announcement-show-start"
                  label={copy(
                    "copy.admin.components.admin.announcements.announcement-dialog.6a9906c79f26"
                  )}
                  value={field.state.value}
                  onChange={field.handleChange}
                />
              )}
            </form.Field>
            <form.Field name="show_end">
              {(field) => (
                <DateTimeControl
                  id="announcement-show-end"
                  label={copy(
                    "copy.admin.components.admin.announcements.announcement-dialog.f50276449943"
                  )}
                  value={field.state.value}
                  onChange={field.handleChange}
                />
              )}
            </form.Field>
          </div>

          {submitError ? (
            <p className="text-sm text-destructive" role="alert">
              {submitError}
            </p>
          ) : null}

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline">
                {copy(
                  "copy.admin.components.admin.announcements.announcement-dialog.2cd0f3be8738"
                )}
              </Button>
            </DialogClose>
            <form.Subscribe selector={(state) => state.isSubmitting}>
              {(formSubmitting) => {
                const submitting = isSaving || formSubmitting
                return (
                  <Button type="submit" disabled={submitting}>
                    {submitting
                      ? copy(
                          "copy.admin.components.admin.announcements.announcement-dialog.a1a1720d3eb5"
                        )
                      : copy(
                          "copy.admin.components.admin.announcements.announcement-dialog.a3030bf8f16d"
                        )}
                  </Button>
                )
              }}
            </form.Subscribe>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
