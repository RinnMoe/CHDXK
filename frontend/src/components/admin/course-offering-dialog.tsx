import { copy } from "@/lib/copy"
import { useEffect, useState } from "react"
import type { ReactNode } from "react"
import { RiAddLine, RiDeleteBinLine } from "@remixicon/react"
import type {
  ManagedCourseOfferingDTO,
  SaveManagedCourseOffering,
} from "@/api/admin-course"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { getErrorMessage, type FormSubmitEvent } from "./admin-utils"
import {
  emptySection,
  draftFromSection,
  sectionCommand,
  sectionFields,
  type SectionDraft,
} from "./course-section-draft"

export function OfferingDialog({
  open,
  semester,
  offering,
  isSaving,
  onOpenChange,
  onSave,
}: {
  open: boolean
  semester: string
  offering: ManagedCourseOfferingDTO | null
  isSaving: boolean
  onOpenChange: (open: boolean) => void
  onSave: (command: SaveManagedCourseOffering) => Promise<void>
}) {
  const [code, setCode] = useState("")
  const [sections, setSections] = useState<SectionDraft[]>([
    { ...emptySection },
  ])
  const [error, setError] = useState("")

  useEffect(() => {
    if (!open) return
    setCode(offering?.code ?? "")
    setSections(
      offering?.sections.length
        ? offering.sections.map(draftFromSection)
        : [{ ...emptySection }]
    )
    setError("")
  }, [open, offering])

  function updateSection(
    index: number,
    field: keyof SectionDraft,
    value: string
  ) {
    setSections((current) =>
      current.map((section, i) =>
        i === index ? { ...section, [field]: value } : section
      )
    )
  }

  async function submit(event: FormSubmitEvent) {
    event.preventDefault()
    setError("")
    try {
      await onSave({
        code: code.trim(),
        semester,
        sections: sections.map(sectionCommand),
      })
      onOpenChange(false)
    } catch (saveError) {
      setError(getErrorMessage(saveError))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-4xl">
        <DialogHeader>
          <DialogTitle>
            {offering
              ? copy(
                  "copy.admin.components.admin.course-management-tab.b9f0f3dbbcd3"
                )
              : copy(
                  "copy.admin.components.admin.course-management-tab.45011c48383a"
                )}
          </DialogTitle>
        </DialogHeader>
        <form className="space-y-4" onSubmit={submit}>
          <div className="grid gap-3 sm:grid-cols-2">
            <Field
              label={copy(
                "copy.admin.components.admin.course-management-tab.d48636ed88e4"
              )}
            >
              <Input
                required
                value={code}
                onChange={(event) => setCode(event.target.value)}
              />
            </Field>
            <Field
              label={copy(
                "copy.admin.components.admin.course-management-tab.53ba710d5f39"
              )}
            >
              <Input value={semester} disabled />
            </Field>
          </div>
          <div className="space-y-3">
            {sections.map((section, index) => (
              <details key={index} open className="rounded-md border p-3">
                <summary className="cursor-pointer font-medium">
                  {section.course_name ||
                    copy(
                      "copy.admin.components.admin.course-management-tab.a401d8927a48"
                    )}{" "}
                  ·{" "}
                  {section.lesson_no ||
                    copy(
                      "copy.admin.components.admin.course-management-tab.c7117390d6f1",
                      { v0: index + 1 }
                    )}
                </summary>
                <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                  {sectionFields.map(([field, label, type]) => (
                    <Field key={field} label={label}>
                      <Input
                        type={type}
                        step={type === "number" ? "any" : undefined}
                        value={section[field]}
                        onChange={(event) =>
                          updateSection(index, field, event.target.value)
                        }
                      />
                    </Field>
                  ))}
                  <Field
                    label={copy(
                      "copy.admin.components.admin.course-management-tab.62559f8067e2"
                    )}
                    className="sm:col-span-2 lg:col-span-3"
                  >
                    <Input
                      value={section.teachers}
                      onChange={(event) =>
                        updateSection(index, "teachers", event.target.value)
                      }
                    />
                  </Field>
                  <Field
                    label={copy(
                      "copy.admin.components.admin.course-management-tab.8e2f12907721"
                    )}
                    className="sm:col-span-2 lg:col-span-3"
                  >
                    <Input
                      value={section.remark}
                      onChange={(event) =>
                        updateSection(index, "remark", event.target.value)
                      }
                    />
                  </Field>
                  <Field
                    label={copy(
                      "copy.admin.components.admin.course-management-tab.e761e8b7e752"
                    )}
                    className="sm:col-span-2 lg:col-span-3"
                  >
                    <Input
                      value={section.syllabus_url}
                      onChange={(event) =>
                        updateSection(index, "syllabus_url", event.target.value)
                      }
                      placeholder={copy(
                        "copy.admin.components.admin.course-offering-dialog.960a47903d29"
                      )}
                    />
                  </Field>
                </div>
                <Button
                  className="mt-3"
                  variant="outline"
                  size="sm"
                  type="button"
                  onClick={() =>
                    setSections((current) =>
                      current.filter((_, i) => i !== index)
                    )
                  }
                  disabled={sections.length <= 1}
                >
                  <RiDeleteBinLine data-icon="inline-start" />
                  {copy(
                    "copy.admin.components.admin.course-management-tab.1f649aa190d6"
                  )}
                </Button>
              </details>
            ))}
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={() =>
              setSections((current) => [...current, { ...emptySection }])
            }
          >
            <RiAddLine data-icon="inline-start" />
            {copy(
              "copy.admin.components.admin.course-management-tab.a25b2c4fe4ab"
            )}
          </Button>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {copy(
                "copy.admin.components.admin.course-management-tab.2cd0f3be8738"
              )}
            </Button>
            <Button type="submit" disabled={isSaving}>
              {isSaving
                ? copy(
                    "copy.admin.components.admin.course-management-tab.ff509c9ba052"
                  )
                : copy(
                    "copy.admin.components.admin.course-management-tab.a3030bf8f16d"
                  )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export function Field({
  label,
  className,
  children,
}: {
  label: string
  className?: string
  children: ReactNode
}) {
  return (
    <label className={`grid gap-1.5 text-sm ${className ?? ""}`}>
      <Label>{label}</Label>
      {children}
    </label>
  )
}
