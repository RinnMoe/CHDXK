import { SitePageContent } from "@/components/common/site-page-content"
import { copy } from "@/lib/copy"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { brand } from "@/config/brand"

export function AboutContent() {
  return <SitePageContent slug="about" />
}

export function AboutAgreementExcerpt() {
  return (
    <Dialog>
      <section className="mx-auto mt-8 max-w-sm text-center text-sm leading-7 text-muted-foreground">
        {copy("copy.user.components.about.about-content.dc616a9a0de1")}
        <DialogTrigger asChild>
          <Button variant="link" className="h-auto px-1 py-0 align-baseline">
            {copy("copy.user.components.about.about-content.e36e41c3d53b")}
          </Button>
        </DialogTrigger>
        。
      </section>

      <DialogContent className="max-h-[calc(100dvh-4rem)] overflow-y-auto sm:max-w-3xl">
        <DialogTitle className="sr-only">
          {copy("copy.user.components.about.about-content.e36e41c3d53b")}
        </DialogTitle>
        <DialogDescription className="sr-only">
          {copy("copy.user.components.about.about-content.52959a13f5c8")}
          {brand.name}
          {copy("copy.user.components.about.about-content.29725de3d959")}
        </DialogDescription>
        <AboutContent />
      </DialogContent>
    </Dialog>
  )
}
