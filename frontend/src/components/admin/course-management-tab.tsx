import { copy } from "@/lib/copy"
import { useEffect, useMemo, useState } from "react"
import {
  RiAddLine,
  RiDeleteBinLine,
  RiEditLine,
  RiUpload2Line,
} from "@remixicon/react"
import type {
  ManagedCourseOfferingDTO,
  SaveManagedCourseOffering,
} from "@/api/admin-course"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  useDeleteManagedCourse,
  useImportManagedCourses,
  useManagedCourses,
  useManagedCourseSemesters,
  useSaveManagedCourse,
} from "@/hooks/use-admin-course"
import { getErrorMessage } from "./admin-utils"

import { Field, OfferingDialog } from "./course-offering-dialog"

export function CourseManagementTab() {
  const semestersQuery = useManagedCourseSemesters()
  const semesters = useMemo(
    () => semestersQuery.data ?? [],
    [semestersQuery.data]
  )
  const [semester, setSemester] = useState("")
  const [query, setQuery] = useState("")
  const [page, setPage] = useState(1)
  const [editing, setEditing] = useState<ManagedCourseOfferingDTO | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [status, setStatus] = useState("")
  const [error, setError] = useState("")
  const courseQuery = useManagedCourses({
    semester,
    q: query,
    page,
    page_size: 50,
  })
  const importMutation = useImportManagedCourses()
  const saveMutation = useSaveManagedCourse()
  const deleteMutation = useDeleteManagedCourse()

  useEffect(() => {
    if (!semester && semesters.length > 0) setSemester(semesters[0])
  }, [semester, semesters])

  function openCreate() {
    setEditing(null)
    setDialogOpen(true)
  }

  async function save(command: SaveManagedCourseOffering) {
    await saveMutation.mutateAsync({ id: editing?.id, command })
    setStatus(
      copy("copy.admin.components.admin.course-management-tab.c0a3807ee203")
    )
  }

  async function upload(file?: File) {
    if (!file) return
    if (
      !window.confirm(
        copy("copy.admin.components.admin.course-management-tab.abacc4e7004d")
      )
    )
      return
    setError("")
    setStatus("")
    try {
      const result = await importMutation.mutateAsync(file)
      setSemester(result.semester)
      setPage(1)
      setStatus(
        copy("copy.admin.components.admin.course-management-tab.f58154bb6b6c", {
          v0: result.semester,
          v1: result.course_count,
          v2: result.section_count,
          v3: result.replaced_count,
        })
      )
    } catch (uploadError) {
      setError(getErrorMessage(uploadError))
    }
  }

  async function remove(offering: ManagedCourseOfferingDTO) {
    if (
      !window.confirm(
        copy("copy.admin.components.admin.course-management-tab.6b6118d68bfa", {
          v0: offering.code,
          v1: offering.semester,
        })
      )
    )
      return
    setError("")
    try {
      await deleteMutation.mutateAsync(offering.id)
      setStatus(
        copy("copy.admin.components.admin.course-management-tab.250828f6f866")
      )
    } catch (deleteError) {
      setError(getErrorMessage(deleteError))
    }
  }

  const pageCount = Math.max(
    1,
    Math.ceil(
      (courseQuery.data?.total ?? 0) / (courseQuery.data?.page_size ?? 50)
    )
  )

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h2 className="text-lg font-medium">
            {copy(
              "copy.admin.components.admin.course-management-tab.f2571f734a5d"
            )}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {copy(
              "copy.admin.components.admin.course-management-tab.441b9c56366f"
            )}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button onClick={openCreate} disabled={!semester}>
            <RiAddLine data-icon="inline-start" />
            {copy(
              "copy.admin.components.admin.course-management-tab.65ce7c3c7b0d"
            )}
          </Button>
          <label className="inline-flex h-9 cursor-pointer items-center gap-2 rounded-md border px-3 text-sm hover:bg-muted/40">
            <RiUpload2Line />
            {importMutation.isPending
              ? copy(
                  "copy.admin.components.admin.course-management-tab.ce1c9d61565f"
                )
              : copy(
                  "copy.admin.components.admin.course-management-tab.505e872bcb8e"
                )}
            <input
              className="sr-only"
              type="file"
              accept=".zip,application/zip"
              disabled={importMutation.isPending}
              onChange={(event) => {
                const file = event.target.files?.[0]
                event.target.value = ""
                void upload(file)
              }}
            />
          </label>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-[minmax(12rem,18rem)_minmax(16rem,1fr)]">
        <Field
          label={copy(
            "copy.admin.components.admin.course-management-tab.2baec43adb17"
          )}
        >
          <Input
            list="managed-course-semesters"
            value={semester}
            onChange={(event) => {
              setSemester(event.target.value)
              setPage(1)
            }}
            placeholder={copy(
              "copy.admin.components.admin.course-management-tab.7028305cf45b"
            )}
          />
          <datalist id="managed-course-semesters">
            {semesters.map((item) => (
              <option key={item} value={item} />
            ))}
          </datalist>
        </Field>
        <Field
          label={copy(
            "copy.admin.components.admin.course-management-tab.8abe5a36638c"
          )}
        >
          <Input
            value={query}
            onChange={(event) => {
              setQuery(event.target.value)
              setPage(1)
            }}
            placeholder={copy(
              "copy.admin.components.admin.course-management-tab.351a38f2a874"
            )}
          />
        </Field>
      </div>

      {status && (
        <p role="status" className="text-sm text-muted-foreground">
          {status}
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}

      {dialogOpen && (
        <OfferingDialog
          open={dialogOpen}
          semester={semester}
          offering={editing}
          isSaving={saveMutation.isPending}
          onOpenChange={setDialogOpen}
          onSave={save}
        />
      )}

      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.course-management-tab.d48636ed88e4"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.course-management-tab.bd2d74b297d1"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.course-management-tab.1e3c672a9fba"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.course-management-tab.6322fa286117"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.course-management-tab.4c9c646f81d1"
                )}
              </TableHead>
              <TableHead className="text-right">
                {copy(
                  "copy.admin.components.admin.course-management-tab.ed31fbb483ee"
                )}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {courseQuery.isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="py-8 text-center text-muted-foreground"
                >
                  {copy(
                    "copy.admin.components.admin.course-management-tab.4927a53bcc88"
                  )}
                </TableCell>
              </TableRow>
            ) : (courseQuery.data?.items ?? []).length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="py-8 text-center text-muted-foreground"
                >
                  {copy(
                    "copy.admin.components.admin.course-management-tab.129336fa8de4"
                  )}
                </TableCell>
              </TableRow>
            ) : (
              courseQuery.data?.items.map((offering) => {
                const section = offering.sections[0]
                return (
                  <TableRow key={offering.id}>
                    <TableCell className="font-mono">{offering.code}</TableCell>
                    <TableCell>
                      {section?.course_name ??
                        copy(
                          "copy.admin.components.admin.course-management-tab.043243bc4b5b"
                        )}
                    </TableCell>
                    <TableCell>
                      {section?.teachers[0] ??
                        copy(
                          "copy.admin.components.admin.course-management-tab.521ea9788171"
                        )}
                    </TableCell>
                    <TableCell>{section?.department ?? "—"}</TableCell>
                    <TableCell>{offering.sections.length}</TableCell>
                    <TableCell className="text-right">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setEditing(offering)
                          setDialogOpen(true)
                        }}
                        aria-label={copy(
                          "copy.admin.components.admin.course-management-tab.3b956ae57c11"
                        )}
                      >
                        <RiEditLine />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="text-destructive"
                        onClick={() => void remove(offering)}
                        aria-label={copy(
                          "copy.admin.components.admin.course-management-tab.92eeff24f15c"
                        )}
                      >
                        <RiDeleteBinLine />
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>
          {copy(
            "copy.admin.components.admin.course-management-tab.2428420e407f"
          )}
          {courseQuery.data?.total ?? 0}
          {copy(
            "copy.admin.components.admin.course-management-tab.88165354d9d3"
          )}
        </span>
        <div className="flex items-center gap-2">
          <span>
            {copy(
              "copy.admin.components.admin.course-management-tab.9552fe3bd366"
            )}
            {page} / {pageCount}
            {copy(
              "copy.admin.components.admin.course-management-tab.85ca81cf6abd"
            )}
          </span>
          <Button
            size="sm"
            variant="outline"
            disabled={page <= 1}
            onClick={() => setPage((value) => Math.max(1, value - 1))}
          >
            {copy(
              "copy.admin.components.admin.course-management-tab.c9b9ae7a6144"
            )}
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={page >= pageCount}
            onClick={() => setPage((value) => Math.min(pageCount, value + 1))}
          >
            {copy(
              "copy.admin.components.admin.course-management-tab.8a8542f69648"
            )}
          </Button>
        </div>
      </div>
    </section>
  )
}
