import { useState } from "react"
import { RiAddLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useCreateCourseEnrollment } from "@/hooks/use-course"

interface CourseEnrollmentDialogProps {
  courseID: number
  courseName: string
  semesters: string[]
}

export function CourseEnrollmentDialog({
  courseID,
  courseName,
  semesters,
}: CourseEnrollmentDialogProps) {
  const [open, setOpen] = useState(false)
  const [semester, setSemester] = useState(semesters[0] ?? "")
  const [error, setError] = useState<string | null>(null)
  const mutation = useCreateCourseEnrollment()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!semester) {
      setError("请选择学期")
      return
    }
    try {
      await mutation.mutateAsync({ courseID, semester })
      setOpen(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : "添加失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm" variant="outline" className="h-8 px-2">
          <RiAddLine data-icon="inline-start" />
          记入选课
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>加入选课记录</DialogTitle>
          <DialogDescription>{courseName}</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-2">
            <Label>学期</Label>
            <Select value={semester} onValueChange={setSemester}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="选择学期" />
              </SelectTrigger>
              <SelectContent>
                {semesters.map((s) => (
                  <SelectItem key={s} value={s}>
                    {s}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {error && (
              <p className="text-sm text-destructive" role="alert">
                {error}
              </p>
            )}
          </div>

          <DialogFooter>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? "添加中" : "添加"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
