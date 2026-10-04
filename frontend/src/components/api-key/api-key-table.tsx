import { copy } from "@/lib/copy"
import { useState } from "react"
import { RiDeleteBinLine } from "@remixicon/react"
import type { ApiKeyDTO } from "@/api/api-key"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatNullableDateTime } from "@/lib/date"
import { LastUsedTableHead } from "./last-used-table-head"

interface ApiKeyTableProps {
  apiKeys: ApiKeyDTO[]
  isLoading: boolean
  isDeleting: boolean
  emptyText: string
  deleteAriaLabel: string
  deleteDialogTitle: string
  deleteDialogDescription: string
  deleteActionLabel: string
  onDelete: (id: string) => Promise<void>
  getDeleteErrorMessage?: (error: unknown) => string
}

function defaultGetDeleteErrorMessage(error: unknown) {
  return error instanceof Error
    ? error.message
    : copy("copy.user.components.api-key.api-key-table.c228558cf257")
}

export function ApiKeyTable({
  apiKeys,
  isLoading,
  isDeleting,
  emptyText,
  deleteAriaLabel,
  deleteDialogTitle,
  deleteDialogDescription,
  deleteActionLabel,
  onDelete,
  getDeleteErrorMessage = defaultGetDeleteErrorMessage,
}: ApiKeyTableProps) {
  const [error, setError] = useState("")

  async function handleDelete(id: string) {
    setError("")
    try {
      await onDelete(id)
    } catch (err) {
      setError(getDeleteErrorMessage(err))
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    )
  }

  if (apiKeys.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        {emptyText}
      </p>
    )
  }

  return (
    <div className="space-y-3">
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>
              {copy("copy.user.components.api-key.api-key-table.d44e9b3d3b31")}
            </TableHead>
            <TableHead>
              {copy("copy.user.components.api-key.api-key-table.99a52df3ff3d")}
            </TableHead>
            <TableHead>
              {copy("copy.user.components.api-key.api-key-table.07ec86e0f1d4")}
            </TableHead>
            <TableHead>
              <LastUsedTableHead />
            </TableHead>
            <TableHead className="w-16 text-right md:w-24">
              {copy("copy.user.components.api-key.api-key-table.ed31fbb483ee")}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {apiKeys.map((key) => (
            <TableRow key={key.id}>
              <TableCell className="font-medium">{key.name}</TableCell>
              <TableCell>
                <code className="rounded bg-muted px-2 py-1 text-xs">
                  {key.key}
                </code>
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatNullableDateTime(key.created_at)}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatNullableDateTime(
                  key.last_used_at,
                  copy(
                    "copy.user.components.api-key.api-key-table.458529e06228"
                  )
                )}
              </TableCell>
              <TableCell className="text-right">
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      className="hover:bg-destructive/10 hover:text-destructive focus-visible:border-destructive/40 focus-visible:ring-destructive/20 dark:hover:bg-destructive/20"
                      aria-label={deleteAriaLabel}
                      disabled={isDeleting}
                    >
                      <RiDeleteBinLine />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent size="sm">
                    <AlertDialogHeader>
                      <AlertDialogTitle>{deleteDialogTitle}</AlertDialogTitle>
                      <AlertDialogDescription>
                        {deleteDialogDescription}
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>
                        {copy(
                          "copy.user.components.api-key.api-key-table.2cd0f3be8738"
                        )}
                      </AlertDialogCancel>
                      <AlertDialogAction
                        variant="destructive"
                        onClick={() => handleDelete(key.id)}
                      >
                        {deleteActionLabel}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
