import { useEffect, useMemo, useState } from "react"
import type { FormEvent, ReactNode } from "react"
import {
	RiAddLine,
	RiDeleteBinLine,
	RiEditLine,
	RiUpload2Line,
} from "@remixicon/react"
import type {
	CourseSectionDTO,
} from "@/api/course"
import type {
	ManagedCourseOfferingDTO,
	SaveManagedCourseOffering,
} from "@/api/admin-course"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import {
	useDeleteManagedCourse,
	useImportManagedCourses,
	useManagedCourses,
	useManagedCourseSemesters,
	useSaveManagedCourse,
} from "@/hooks/use-admin-course"
import { getErrorMessage } from "./admin-utils"

interface SectionDraft {
	source_lesson_id: string
	lesson_no: string
	course_name: string
	course_type: string
	credit: string
	department: string
	language: string
	campus: string
	assessment: string
	teach_class: string
	actual_count: string
	limit_count: string
	course_period: string
	total_hours: string
	weekly_hours: string
	week_range: string
	start_week: string
	end_week: string
	week_count: string
	remark: string
	syllabus_url: string
	teachers: string
}

const emptySection: SectionDraft = {
	source_lesson_id: "",
	lesson_no: "",
	course_name: "",
	course_type: "",
	credit: "0",
	department: "",
	language: "",
	campus: "",
	assessment: "",
	teach_class: "",
	actual_count: "",
	limit_count: "",
	course_period: "",
	total_hours: "",
	weekly_hours: "",
	week_range: "",
	start_week: "",
	end_week: "",
	week_count: "",
	remark: "",
	syllabus_url: "",
	teachers: "",
}

function draftFromSection(section?: CourseSectionDTO): SectionDraft {
	if (!section) return { ...emptySection }
	return {
		source_lesson_id: section.source_lesson_id ?? "",
		lesson_no: section.lesson_no ?? "",
		course_name: section.course_name ?? "",
		course_type: section.course_type ?? "",
		credit: String(section.credit ?? 0),
		department: section.department ?? "",
		language: section.language ?? "",
		campus: section.campus ?? "",
		assessment: section.assessment ?? "",
		teach_class: section.teach_class ?? "",
		actual_count: section.actual_count == null ? "" : String(section.actual_count),
		limit_count: section.limit_count == null ? "" : String(section.limit_count),
		course_period: section.course_period ?? "",
		total_hours: section.total_hours == null ? "" : String(section.total_hours),
		weekly_hours: section.weekly_hours == null ? "" : String(section.weekly_hours),
		week_range: section.week_range ?? "",
		start_week: section.start_week == null ? "" : String(section.start_week),
		end_week: section.end_week == null ? "" : String(section.end_week),
		week_count: section.week_count == null ? "" : String(section.week_count),
		remark: section.remark ?? "",
		syllabus_url: section.syllabus_url ?? "",
		teachers: (section.teachers ?? []).join("，"),
	}
}

function optionalNumber(value: string, field: string) {
	if (!value.trim()) return null
	const number = Number(value)
	if (!Number.isFinite(number)) throw new Error(`${field} 必须是有效数字`)
	return number
}

function sectionCommand(section: SectionDraft) {
	const credit = optionalNumber(section.credit, "学分")
	if (credit == null) throw new Error("学分不能为空")
	return {
		source_lesson_id: section.source_lesson_id.trim(),
		lesson_no: section.lesson_no.trim(),
		course_name: section.course_name.trim(),
		course_type: section.course_type.trim(),
		credit,
		department: section.department.trim(),
		language: section.language.trim(),
		campus: section.campus.trim(),
		assessment: section.assessment.trim(),
		teach_class: section.teach_class.trim(),
		actual_count: optionalNumber(section.actual_count, "实际人数"),
		limit_count: optionalNumber(section.limit_count, "人数上限"),
		course_period: section.course_period.trim(),
		total_hours: optionalNumber(section.total_hours, "总课时"),
		weekly_hours: optionalNumber(section.weekly_hours, "周课时"),
		week_range: section.week_range.trim(),
		start_week: optionalNumber(section.start_week, "起始周"),
		end_week: optionalNumber(section.end_week, "结束周"),
		week_count: optionalNumber(section.week_count, "周数"),
		remark: section.remark.trim(),
		syllabus_url: section.syllabus_url.trim(),
		teachers: section.teachers.split(/[，,;；]/).map((name) => name.trim()).filter(Boolean),
	}
}

function OfferingDialog({
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
	const [sections, setSections] = useState<SectionDraft[]>([{ ...emptySection }])
	const [error, setError] = useState("")

	useEffect(() => {
		if (!open) return
		setCode(offering?.code ?? "")
		setSections(offering?.sections.length ? offering.sections.map(draftFromSection) : [{ ...emptySection }])
		setError("")
	}, [open, offering])

	function updateSection(index: number, field: keyof SectionDraft, value: string) {
		setSections((current) => current.map((section, i) => i === index ? { ...section, [field]: value } : section))
	}

	async function submit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault()
		setError("")
		try {
			await onSave({ code: code.trim(), semester, sections: sections.map(sectionCommand) })
			onOpenChange(false)
		} catch (saveError) {
			setError(getErrorMessage(saveError))
		}
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-4xl">
				<DialogHeader>
					<DialogTitle>{offering ? "编辑学期开课" : "新增学期开课"}</DialogTitle>
				</DialogHeader>
				<form className="space-y-4" onSubmit={submit}>
					<div className="grid gap-3 sm:grid-cols-2">
						<Field label="课程代码">
							<Input required value={code} onChange={(event) => setCode(event.target.value)} />
						</Field>
						<Field label="所属学期">
							<Input value={semester} disabled />
						</Field>
					</div>
					<div className="space-y-3">
						{sections.map((section, index) => (
							<details key={index} open className="rounded-md border p-3">
								<summary className="cursor-pointer font-medium">
									{section.course_name || "新教学班"} · {section.lesson_no || `教学班 ${index + 1}`}
								</summary>
								<div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
									{sectionFields.map(([field, label, type]) => (
										<Field key={field} label={label}>
											<Input
												type={type}
												step={type === "number" ? "any" : undefined}
												value={section[field]}
												onChange={(event) => updateSection(index, field, event.target.value)}
											/>
										</Field>
									))}
									<Field label="教师（按主讲、合讲顺序，用逗号分隔）" className="sm:col-span-2 lg:col-span-3">
										<Input value={section.teachers} onChange={(event) => updateSection(index, "teachers", event.target.value)} />
									</Field>
									<Field label="课程备注" className="sm:col-span-2 lg:col-span-3">
										<Input value={section.remark} onChange={(event) => updateSection(index, "remark", event.target.value)} />
									</Field>
									<Field label="CHD 大纲链接" className="sm:col-span-2 lg:col-span-3">
										<Input value={section.syllabus_url} onChange={(event) => updateSection(index, "syllabus_url", event.target.value)} placeholder="https://bkjw.chd.edu.cn/eams/..." />
									</Field>
								</div>
								<Button
									className="mt-3"
									variant="outline"
									size="sm"
									type="button"
									onClick={() => setSections((current) => current.filter((_, i) => i !== index))}
									disabled={sections.length <= 1}
								>
									<RiDeleteBinLine data-icon="inline-start" />
									删除教学班
								</Button>
							</details>
						))}
					</div>
					<Button type="button" variant="outline" onClick={() => setSections((current) => [...current, { ...emptySection }])}>
						<RiAddLine data-icon="inline-start" />添加教学班
					</Button>
					{error && <p className="text-sm text-destructive">{error}</p>}
					<DialogFooter>
						<Button type="button" variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
						<Button type="submit" disabled={isSaving}>{isSaving ? "保存中…" : "保存"}</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	)
}

const sectionFields: [keyof SectionDraft, string, "text" | "number"][] = [
	["lesson_no", "课程序号", "text"],
	["source_lesson_id", "来源课节 ID", "text"],
	["course_name", "课程名称", "text"],
	["course_type", "课程类别", "text"],
	["credit", "学分", "number"],
	["department", "开课院系", "text"],
	["language", "授课语言", "text"],
	["campus", "校区", "text"],
	["assessment", "考核方式", "text"],
	["teach_class", "教学班", "text"],
	["actual_count", "实际人数", "number"],
	["limit_count", "人数上限", "number"],
	["course_period", "课程周期", "text"],
	["total_hours", "总课时", "number"],
	["weekly_hours", "周课时", "number"],
	["week_range", "起止周", "text"],
	["start_week", "起始周", "number"],
	["end_week", "结束周", "number"],
	["week_count", "周数", "number"],
]

function Field({ label, className, children }: { label: string; className?: string; children: ReactNode }) {
	return <label className={`grid gap-1.5 text-sm ${className ?? ""}`}><Label>{label}</Label>{children}</label>
}

export function CourseManagementTab() {
	const semestersQuery = useManagedCourseSemesters()
	const semesters = useMemo(() => semestersQuery.data ?? [], [semestersQuery.data])
	const [semester, setSemester] = useState("")
	const [query, setQuery] = useState("")
	const [page, setPage] = useState(1)
	const [editing, setEditing] = useState<ManagedCourseOfferingDTO | null>(null)
	const [dialogOpen, setDialogOpen] = useState(false)
	const [status, setStatus] = useState("")
	const [error, setError] = useState("")
	const courseQuery = useManagedCourses({ semester, q: query, page, page_size: 50 })
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
		setStatus("课程信息已保存")
	}

	async function upload(file?: File) {
		if (!file) return
		if (!window.confirm("重新上传会完整替换所选学期的课程和教学班；历史课程评价会保留。继续吗？")) return
		setError("")
		setStatus("")
		try {
			const result = await importMutation.mutateAsync(file)
			setSemester(result.semester)
			setPage(1)
			setStatus(`已导入 ${result.semester}：${result.course_count} 门课程、${result.section_count} 个教学班；替换 ${result.replaced_count} 条旧开课记录。`)
		} catch (uploadError) {
			setError(getErrorMessage(uploadError))
		}
	}

	async function remove(offering: ManagedCourseOfferingDTO) {
		if (!window.confirm(`删除 ${offering.code} 在 ${offering.semester} 的开课记录？历史评价会保留。`)) return
		setError("")
		try {
			await deleteMutation.mutateAsync(offering.id)
			setStatus("学期开课记录已删除，历史评价已保留")
		} catch (deleteError) {
			setError(getErrorMessage(deleteError))
		}
	}

	const pageCount = Math.max(1, Math.ceil((courseQuery.data?.total ?? 0) / (courseQuery.data?.page_size ?? 50)))

	return (
		<section className="space-y-4">
			<div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<h2 className="text-lg font-medium">课程与学期数据</h2>
					<p className="mt-1 text-sm text-muted-foreground">上传完整 ZIP 会替换该学期的课程和教学班；已发布的课程评价保留。</p>
				</div>
				<div className="flex flex-wrap items-center gap-2">
					<Button onClick={openCreate} disabled={!semester}>
						<RiAddLine data-icon="inline-start" />新增课程
					</Button>
					<label className="inline-flex h-9 cursor-pointer items-center gap-2 rounded-md border px-3 text-sm hover:bg-muted/40">
						<RiUpload2Line />
						{importMutation.isPending ? "上传中…" : "上传学期 ZIP"}
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
				<Field label="按学期管理">
					<Input list="managed-course-semesters" value={semester} onChange={(event) => { setSemester(event.target.value); setPage(1) }} placeholder="选择或输入学期，如 2026-2027-1" />
					<datalist id="managed-course-semesters">
						{semesters.map((item) => <option key={item} value={item} />)}
					</datalist>
				</Field>
				<Field label="搜索课程代码、名称或教学班">
					<Input value={query} onChange={(event) => { setQuery(event.target.value); setPage(1) }} placeholder="输入课程代码、名称或课程序号" />
				</Field>
			</div>

			{status && <p role="status" className="text-sm text-muted-foreground">{status}</p>}
			{error && <p role="alert" className="text-sm text-destructive">{error}</p>}

			{dialogOpen && <OfferingDialog
				open={dialogOpen}
				semester={semester}
				offering={editing}
				isSaving={saveMutation.isPending}
				onOpenChange={setDialogOpen}
				onSave={save}
			/>}

			<div className="overflow-x-auto rounded-lg border">
				<Table>
					<TableHeader><TableRow><TableHead>课程代码</TableHead><TableHead>课程名称</TableHead><TableHead>主讲教师</TableHead><TableHead>开课单位</TableHead><TableHead>教学班</TableHead><TableHead className="text-right">操作</TableHead></TableRow></TableHeader>
					<TableBody>
						{courseQuery.isLoading ? <TableRow><TableCell colSpan={6} className="py-8 text-center text-muted-foreground">加载中…</TableCell></TableRow>
							: (courseQuery.data?.items ?? []).length === 0 ? <TableRow><TableCell colSpan={6} className="py-8 text-center text-muted-foreground">该学期暂无课程数据</TableCell></TableRow>
								: courseQuery.data?.items.map((offering) => {
									const section = offering.sections[0]
									return <TableRow key={offering.id}>
										<TableCell className="font-mono">{offering.code}</TableCell>
										<TableCell>{section?.course_name ?? "课程信息暂缺"}</TableCell>
										<TableCell>{section?.teachers[0] ?? "教师信息暂缺"}</TableCell>
										<TableCell>{section?.department ?? "—"}</TableCell>
										<TableCell>{offering.sections.length}</TableCell>
										<TableCell className="text-right">
											<Button size="sm" variant="ghost" onClick={() => { setEditing(offering); setDialogOpen(true) }} aria-label="编辑课程"><RiEditLine /></Button>
											<Button size="sm" variant="ghost" className="text-destructive" onClick={() => void remove(offering)} aria-label="删除学期开课"><RiDeleteBinLine /></Button>
										</TableCell>
									</TableRow>
								})}
					</TableBody>
				</Table>
			</div>
			<div className="flex items-center justify-between text-sm text-muted-foreground">
				<span>共 {courseQuery.data?.total ?? 0} 门课程</span>
				<div className="flex items-center gap-2">
					<span>第 {page} / {pageCount} 页</span>
					<Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((value) => Math.max(1, value - 1))}>上一页</Button>
					<Button size="sm" variant="outline" disabled={page >= pageCount} onClick={() => setPage((value) => Math.min(pageCount, value + 1))}>下一页</Button>
				</div>
			</div>
		</section>
	)
}
