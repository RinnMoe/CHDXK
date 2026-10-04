import { copy } from "@/lib/copy"
import { useState } from "react"
import { RiAddLine } from "@remixicon/react"
import type {
  AnnouncementDTO,
  SaveAnnouncementCommand,
} from "@/api/announcement"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TooltipProvider } from "@/components/ui/tooltip"
import {
  useAdminAnnouncements,
  useCreateAnnouncement,
  useDeleteAnnouncement,
  useUpdateAnnouncement,
} from "@/hooks/use-announcement"
import { AnnouncementDialog } from "./announcements/announcement-dialog"
import { AnnouncementRow } from "./announcements/announcement-row"

export function AnnouncementsTab() {
  const announcementsQuery = useAdminAnnouncements()
  const createMutation = useCreateAnnouncement()
  const updateMutation = useUpdateAnnouncement()
  const deleteMutation = useDeleteAnnouncement()
  const [editing, setEditing] = useState<AnnouncementDTO | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)

  function openCreate() {
    setEditing(null)
    setDialogOpen(true)
  }

  function openEdit(announcement: AnnouncementDTO) {
    setEditing(announcement)
    setDialogOpen(true)
  }

  async function saveAnnouncement(cmd: SaveAnnouncementCommand) {
    if (editing) {
      await updateMutation.mutateAsync({ id: editing.id, cmd })
    } else {
      await createMutation.mutateAsync(cmd)
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <h2 className="text-lg font-medium">
          {copy("copy.admin.components.admin.announcements-tab.069c0ecbf898")}
        </h2>
        <Button onClick={openCreate}>
          <RiAddLine data-icon="inline-start" />
          {copy("copy.admin.components.admin.announcements-tab.ee4ea097b207")}
        </Button>
      </div>

      {dialogOpen ? (
        <AnnouncementDialog
          open={dialogOpen}
          announcement={editing}
          isSaving={createMutation.isPending || updateMutation.isPending}
          onOpenChange={setDialogOpen}
          onSave={saveAnnouncement}
        />
      ) : null}

      <TooltipProvider>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.announcements-tab.3843971dcfde"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.announcements-tab.c3405f8c7d9d"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.announcements-tab.565d64601d4d"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.announcements-tab.1ac47221096c"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.announcements-tab.05e99264a759"
                )}
              </TableHead>
              <TableHead>
                {copy(
                  "copy.admin.components.admin.announcements-tab.07ec86e0f1d4"
                )}
              </TableHead>
              <TableHead className="text-right">
                {copy(
                  "copy.admin.components.admin.announcements-tab.ed31fbb483ee"
                )}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {announcementsQuery.isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="py-8 text-center text-muted-foreground"
                >
                  {copy(
                    "copy.admin.components.admin.announcements-tab.d04fcbda737f"
                  )}
                </TableCell>
              </TableRow>
            ) : (announcementsQuery.data ?? []).length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="py-8 text-center text-muted-foreground"
                >
                  {copy(
                    "copy.admin.components.admin.announcements-tab.6d9a7758610a"
                  )}
                </TableCell>
              </TableRow>
            ) : (
              (announcementsQuery.data ?? []).map((announcement) => (
                <AnnouncementRow
                  key={announcement.id}
                  announcement={announcement}
                  isDeleting={deleteMutation.isPending}
                  onEdit={openEdit}
                  onDelete={(id) => {
                    void deleteMutation.mutateAsync(id)
                  }}
                />
              ))
            )}
          </TableBody>
        </Table>
      </TooltipProvider>
    </section>
  )
}
