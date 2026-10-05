import { copy } from "@/lib/copy"

import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import {
  useAdminSystemSettings,
  useUpdateSystemSetting,
} from "@/hooks/use-system-settings"
import type { SitePage } from "@/api/site-pages"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { brand } from "@/config/brand"
import { SafeMarkdown } from "@/components/review/safe-markdown"
import { getErrorMessage } from "./admin-utils"

function PageEditor({
  page,
  onSave,
  onCancel,
  pending,
}: {
  page: SitePage
  onSave: (page: SitePage) => void
  onCancel: () => void
  pending: boolean
}) {
  const [draft, setDraft] = useState(page)
  const [preview, setPreview] = useState(false)
  return (
    <div className="space-y-4 rounded-lg border p-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="page-title">
            {copy(
              "copy.admin.components.admin.content-pages-panel.5ce2d4cc8539"
            )}
          </Label>
          <Input
            id="page-title"
            value={draft.title}
            onChange={(e) => setDraft({ ...draft, title: e.target.value })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="page-slug">
            {copy(
              "copy.admin.components.admin.content-pages-panel.fff67cc3d883"
            )}
          </Label>
          <Input
            id="page-slug"
            value={draft.slug}
            onChange={(e) => setDraft({ ...draft, slug: e.target.value })}
          />
          <p className="text-xs text-muted-foreground">
            {copy(
              "copy.admin.components.admin.content-pages-panel.0e8198de6be9"
            )}
          </p>
        </div>
      </div>
      <div className="flex flex-wrap gap-4">
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={draft.published}
            onChange={(e) =>
              setDraft({ ...draft, published: e.target.checked })
            }
          />
          {copy("copy.admin.components.admin.content-pages-panel.5c5a53a77586")}
        </label>
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={draft.footer}
            onChange={(e) => setDraft({ ...draft, footer: e.target.checked })}
          />
          {copy("copy.admin.components.admin.content-pages-panel.cc38b219f83b")}
        </label>
      </div>
      <div className="flex items-center justify-between">
        <Label htmlFor="page-content">
          {copy("copy.admin.components.admin.content-pages-panel.37b5c7a9b366")}
        </Label>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setPreview(!preview)}
        >
          {preview
            ? copy(
                "copy.admin.components.admin.content-pages-panel.fd4b9e3b6c68"
              )
            : copy(
                "copy.admin.components.admin.content-pages-panel.13d61fea9f17"
              )}
        </Button>
      </div>
      {preview ? (
        <article className="prose max-w-none dark:prose-invert">
          <SafeMarkdown
            content={draft.content.replaceAll("{site_name}", brand.name)}
          />
        </article>
      ) : (
        <Textarea
          id="page-content"
          className="min-h-96 font-mono"
          value={draft.content}
          onChange={(e) => setDraft({ ...draft, content: e.target.value })}
        />
      )}
      <p className="text-xs text-muted-foreground">
        {copy("copy.admin.components.admin.content-pages-panel.4307781b7e6e")}
        {"{site_name}"}
        {copy("copy.admin.components.admin.content-pages-panel.b21cf74664d4")}
      </p>
      <div className="flex gap-2">
        <Button disabled={pending} onClick={() => onSave(draft)}>
          {pending
            ? copy(
                "copy.admin.components.admin.content-pages-panel.ff509c9ba052"
              )
            : copy(
                "copy.admin.components.admin.content-pages-panel.87817c6cfc6d"
              )}
        </Button>
        <Button variant="outline" disabled={pending} onClick={onCancel}>
          {copy("copy.admin.components.admin.content-pages-panel.2cd0f3be8738")}
        </Button>
      </div>
    </div>
  )
}

export function ContentPagesPanel() {
  const query = useAdminSystemSettings()
  const mutation = useUpdateSystemSetting()
  const qc = useQueryClient()
  const [selected, setSelected] = useState<SitePage | null>(null)
  const [error, setError] = useState("")
  const entry = query.data?.find((item) => item.key === "content.pages")
  const pages: SitePage[] = entry ? JSON.parse(entry.value) : []
  async function save(page: SitePage) {
    setError("")
    try {
      const next = selected?.slug
        ? pages.map((item) => (item.slug === selected.slug ? page : item))
        : [...pages, page]
      await mutation.mutateAsync({
        key: "content.pages",
        cmd: { value: JSON.stringify(next) },
      })
      await qc.invalidateQueries({ queryKey: ["site-pages"] })
      setSelected(null)
    } catch (err) {
      setError(getErrorMessage(err))
    }
  }
  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium">
          {copy("copy.admin.components.admin.content-pages-panel.f52222e9c24c")}
        </h2>
        <Button
          disabled={query.isLoading}
          onClick={() => {
            setError("")
            setSelected({
              slug: "",
              title: "",
              content: "",
              published: false,
              footer: true,
            })
          }}
        >
          {copy("copy.admin.components.admin.content-pages-panel.51a821f2d9eb")}
        </Button>
      </div>
      <p className="text-sm text-muted-foreground">
        {copy("copy.admin.components.admin.content-pages-panel.15ee890b39e9")}
      </p>
      {error && (
        <p role="alert" className="text-destructive">
          {error}
        </p>
      )}
      {query.isLoading ? (
        <p>
          {copy("copy.admin.components.admin.content-pages-panel.4927a53bcc88")}
        </p>
      ) : query.isError ? (
        <p role="alert">
          {copy("copy.admin.components.admin.content-pages-panel.7439a91a2665")}
        </p>
      ) : selected ? (
        <PageEditor
          key={selected.slug}
          page={selected}
          pending={mutation.isPending}
          onSave={(page) => void save(page)}
          onCancel={() => setSelected(null)}
        />
      ) : (
        <div className="divide-y rounded-lg border">
          {pages.map((page) => (
            <div
              key={page.slug}
              className="flex items-center justify-between gap-3 p-4"
            >
              <div>
                <h3 className="font-medium">{page.title}</h3>
                <p className="text-sm text-muted-foreground">
                  {copy(
                    "copy.admin.components.admin.content-pages-panel.b899364faf74"
                  )}
                  {page.slug} ·{" "}
                  {page.published
                    ? copy(
                        "copy.admin.components.admin.content-pages-panel.d334724d6b2f"
                      )
                    : copy(
                        "copy.admin.components.admin.content-pages-panel.2a2fd29bd27a"
                      )}
                  {page.footer
                    ? copy(
                        "copy.admin.components.admin.content-pages-panel.110bf1de7a90"
                      )
                    : ""}
                </p>
              </div>
              <Button variant="outline" onClick={() => setSelected(page)}>
                {copy(
                  "copy.admin.components.admin.content-pages-panel.1d779d74e0b5"
                )}
              </Button>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
