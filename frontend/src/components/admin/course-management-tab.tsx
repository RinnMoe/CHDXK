import { copy } from "@/lib/copy"
import { useMemo, useState } from "react"
import { RiAddLine, RiDeleteBinLine, RiEditLine } from "@remixicon/react"
import type {
  ManagedCourseOfferingDTO,
  SaveManagedCourseOffering,
} from "@/api/admin-course"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableRow } from "@/components/ui/table"
import {
  useDeleteManagedCourse,
  useManagedCourses,
  useManagedCourseSemesters,
  useSaveManagedCourse,
} from "@/hooks/use-admin-course"
import { getErrorMessage } from "./admin-utils"

import {
  ManagedCourseTableHeader,
  ManagedCoursePagination,
} from "./course-management-table-header"
import { Field, OfferingDialog } from "./course-offering-dialog"

export function CourseManagementTab() {
  const semestersQuery = useManagedCourseSemesters()
  const semesters = useMemo(
    () => semestersQuery.data ?? [],
    [semestersQuery.data]
  )
  const [selectedSemester, setSemester] = useState("")
  const semester = selectedSemester || semesters[0] || ""
  const [filters, setFilters] = useState({
    code: "",
    name: "",
    teacher: "",
    department: "",
    lesson_no: "",
    min_sections: "",
    max_sections: "",
  })
  const [sort, setSort] = useState<
    "code" | "name" | "teacher" | "department" | "sections"
  >("code")
  const [direction, setDirection] = useState<"asc" | "desc">("asc")
  const [pageSize, setPageSize] = useState(50)
  const rangeInvalid =
    (filters.min_sections !== "" &&
      (!Number.isInteger(Number(filters.min_sections)) ||
        Number(filters.min_sections) < 0)) ||
    (filters.max_sections !== "" &&
      (!Number.isInteger(Number(filters.max_sections)) ||
        Number(filters.max_sections) < 0)) ||
    (filters.min_sections !== "" &&
      filters.max_sections !== "" &&
      Number(filters.min_sections) > Number(filters.max_sections))
  function updateFilter(key: keyof typeof filters, value: string) {
    setFilters((prev) => ({ ...prev, [key]: value }))
    setPage(1)
  }
  function changeSort(key: typeof sort) {
    setSort(key)
    setDirection(sort === key && direction === "asc" ? "desc" : "asc")
    setPage(1)
  }
  const [page, setPage] = useState(1)
  const [editing, setEditing] = useState<ManagedCourseOfferingDTO | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [status, setStatus] = useState("")
  const [error, setError] = useState("")
  const courseQuery = useManagedCourses(
    {
      semester,
      ...filters,
      min_sections:
        filters.min_sections === "" ? undefined : Number(filters.min_sections),
      max_sections:
        filters.max_sections === "" ? undefined : Number(filters.max_sections),
      sort,
      direction,
      page,
      page_size: pageSize,
    },
    !rangeInvalid
  )
  const saveMutation = useSaveManagedCourse()
  const deleteMutation = useDeleteManagedCourse()

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
      (courseQuery.data?.total ?? 0) / (courseQuery.data?.page_size ?? pageSize)
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
        </div>
      </div>

      <div className="flex flex-wrap items-end gap-3">
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
        <Field label="课程序号">
          <Input
            aria-label="筛选课程序号"
            value={filters.lesson_no}
            onChange={(event) => updateFilter("lesson_no", event.target.value)}
            placeholder="输入课程序号"
          />
        </Field>
        <Field label="每页数量">
          <select
            aria-label="每页数量"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={pageSize}
            onChange={(event) => {
              setPageSize(Number(event.target.value))
              setPage(1)
            }}
          >
            {[20, 50, 100].map((size) => (
              <option key={size} value={size}>
                {size} 条
              </option>
            ))}
          </select>
        </Field>
        <Button
          variant="outline"
          onClick={() => {
            setFilters({
              code: "",
              name: "",
              teacher: "",
              department: "",
              lesson_no: "",
              min_sections: "",
              max_sections: "",
            })
            setSort("code")
            setDirection("asc")
            setPage(1)
          }}
        >
          清除筛选
        </Button>
      </div>

      {status && (
        <p role="status" className="text-sm text-muted-foreground">
          {status}
        </p>
      )}
      {rangeInvalid && (
        <p role="alert" className="text-sm text-destructive">
          教学班数量须为非负整数，最少数量不能大于最多数量。
        </p>
      )}
      {courseQuery.isError && (
        <p role="alert" className="text-sm text-destructive">
          {getErrorMessage(courseQuery.error)}
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
          <ManagedCourseTableHeader
            filters={filters}
            sort={sort}
            direction={direction}
            onFilterChange={updateFilter}
            onSortChange={changeSort}
          />
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
                const courseNames = [
                  ...new Set(
                    offering.sections
                      .map((item) => item.course_name)
                      .filter(Boolean)
                  ),
                ]
                const primaryTeachers = [
                  ...new Set(
                    offering.sections
                      .map((item) => item.teachers[0])
                      .filter(Boolean)
                  ),
                ]
                const departments = [
                  ...new Set(
                    offering.sections
                      .map((item) => item.department)
                      .filter(Boolean)
                  ),
                ]
                return (
                  <TableRow key={offering.id}>
                    <TableCell className="font-mono">{offering.code}</TableCell>
                    <TableCell>
                      {(courseNames.join("、") || undefined) ??
                        copy(
                          "copy.admin.components.admin.course-management-tab.043243bc4b5b"
                        )}
                    </TableCell>
                    <TableCell>
                      {(primaryTeachers.join("、") || undefined) ??
                        copy(
                          "copy.admin.components.admin.course-management-tab.521ea9788171"
                        )}
                    </TableCell>
                    <TableCell>{departments.join("、") || "—"}</TableCell>
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
      <ManagedCoursePagination
        total={courseQuery.data?.total ?? 0}
        page={page}
        pageCount={pageCount}
        onPageChange={setPage}
      />
    </section>
  )
}
