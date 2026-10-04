import { copy } from "@/lib/copy"
import { useState, type SyntheticEvent } from "react"
import { RiFileCopyLine } from "@remixicon/react"
import type { ApiKeyDTO, CreateApiKeyCommand } from "@/api/api-key"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface ApiKeyCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreate: (cmd: CreateApiKeyCommand) => Promise<ApiKeyDTO>
  isCreating: boolean
  title: string
  description: string
  nameInputID: string
  namePlaceholder: string
  getErrorMessage?: (error: unknown) => string
}

function defaultGetErrorMessage(error: unknown) {
  return error instanceof Error
    ? error.message
    : copy("copy.user.components.api-key.api-key-create-dialog.7e6a71efbf63")
}

export function ApiKeyCreateDialog({
  open,
  onOpenChange,
  onCreate,
  isCreating,
  title,
  description,
  nameInputID,
  namePlaceholder,
  getErrorMessage = defaultGetErrorMessage,
}: ApiKeyCreateDialogProps) {
  const [name, setName] = useState("")
  const [createdKey, setCreatedKey] = useState<ApiKeyDTO | null>(null)
  const [copyMessage, setCopyMessage] = useState("")
  const [error, setError] = useState("")

  async function copyKey(value: string) {
    setCopyMessage("")
    try {
      await navigator.clipboard.writeText(value)
      setCopyMessage(
        copy("copy.user.components.api-key.api-key-create-dialog.8f6f8d979c98")
      )
    } catch {
      setCopyMessage(
        copy("copy.user.components.api-key.api-key-create-dialog.753d8bb0da99")
      )
    }
  }

  function handleOpenChange(nextOpen: boolean) {
    onOpenChange(nextOpen)
    if (!nextOpen) {
      setName("")
      setCreatedKey(null)
      setCopyMessage("")
      setError("")
    }
  }

  async function handleCreate(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = name.trim()
    if (!trimmed) {
      setError(
        copy("copy.user.components.api-key.api-key-create-dialog.effa003332b6")
      )
      return
    }

    setError("")
    setCopyMessage("")
    try {
      const key = await onCreate({ name: trimmed })
      setCreatedKey(key)
      setName("")
    } catch (err) {
      setError(getErrorMessage(err))
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        {createdKey?.key ? (
          <div className="space-y-4">
            <div className="flex flex-wrap items-start gap-3">
              <code className="min-w-0 flex-1 rounded-md border bg-muted/30 px-3 py-2 text-sm break-all">
                {createdKey.key}
              </code>
              <Button
                type="button"
                variant="outline"
                onClick={() => copyKey(createdKey.key)}
              >
                <RiFileCopyLine data-icon="inline-start" />
                {copy(
                  "copy.user.components.api-key.api-key-create-dialog.63d90d977348"
                )}
              </Button>
            </div>
            {copyMessage ? (
              <p className="text-sm text-muted-foreground">{copyMessage}</p>
            ) : null}
            <DialogFooter>
              <DialogClose asChild>
                <Button>
                  {copy(
                    "copy.user.components.api-key.api-key-create-dialog.c0b3fbff51cc"
                  )}
                </Button>
              </DialogClose>
            </DialogFooter>
          </div>
        ) : (
          <form className="space-y-4" onSubmit={handleCreate}>
            <div className="space-y-2">
              <Label htmlFor={nameInputID}>
                {copy(
                  "copy.user.components.api-key.api-key-create-dialog.d44e9b3d3b31"
                )}
              </Label>
              <Input
                id={nameInputID}
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder={namePlaceholder}
                maxLength={80}
              />
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="outline">
                  {copy(
                    "copy.user.components.api-key.api-key-create-dialog.2cd0f3be8738"
                  )}
                </Button>
              </DialogClose>
              <Button type="submit" disabled={isCreating}>
                {isCreating
                  ? copy(
                      "copy.user.components.api-key.api-key-create-dialog.67531c9da3b4"
                    )
                  : copy(
                      "copy.user.components.api-key.api-key-create-dialog.cde2cd071d25"
                    )}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
