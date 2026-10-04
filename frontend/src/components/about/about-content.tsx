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

const agreementSections = [
  {
    title: copy("copy.user.components.about.about-content.d0f3e3298b26"),
    paragraphs: [
      <>
        {brand.name}
        {copy("copy.user.components.about.about-content.5c3c998c13cc")}
      </>,
      <>
        {brand.name}
        {copy("copy.user.components.about.about-content.0555414e3117")}
      </>,
    ],
  },
  {
    title: copy("copy.user.components.about.about-content.b4d4f455a802"),
    paragraphs: [
      <>
        {copy("copy.user.components.about.about-content.84b13d2f7676")}
        {brand.name}
        {copy("copy.user.components.about.about-content.0949730fb040")}
      </>,
      <>
        {brand.name}
        {copy("copy.user.components.about.about-content.53037c66add5")}
      </>,
    ],
  },
  {
    title: copy("copy.user.components.about.about-content.86651d17a401"),
    paragraphs: [
      <>
        {brand.name}
        {copy("copy.user.components.about.about-content.98d673965b4e")}
      </>,
      <>{copy("copy.user.components.about.about-content.86e835668ee2")}</>,
      <>{copy("copy.user.components.about.about-content.bbcc08c43c0e")}</>,
    ],
  },
]

function AgreementSections({ compact = false }: { compact?: boolean }) {
  return (
    <>
      {agreementSections.map((section) => (
        <section key={section.title} className="space-y-3">
          <h2
            className={
              compact ? "text-base font-semibold" : "text-lg font-semibold"
            }
          >
            {section.title}
          </h2>
          <div className="space-y-2 text-sm leading-7 text-muted-foreground">
            {section.paragraphs.map((paragraph, index) => (
              <p key={index}>{paragraph}</p>
            ))}
          </div>
        </section>
      ))}
    </>
  )
}

export function AboutContent() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">
          {copy("copy.user.components.about.about-content.e499019ef2e3")}
          {brand.name}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {brand.name}
          {copy("copy.user.components.about.about-content.25f8c737e369")}
        </p>
      </div>

      <AgreementSections />

      {brand.feedbackEmail && (
        <section className="space-y-3">
          <h2 className="text-lg font-semibold">
            {copy("copy.user.components.about.about-content.323a881ebd47")}
          </h2>
          <p className="text-sm leading-7 text-muted-foreground">
            {copy("copy.user.components.about.about-content.1d76ac77ffe8")}{" "}
            <a
              href={`mailto:${brand.feedbackEmail}`}
              className="font-medium text-primary hover:underline"
            >
              {brand.feedbackEmail}
            </a>{" "}
            {copy("copy.user.components.about.about-content.c4622260f7c7")}
          </p>
        </section>
      )}
    </div>
  )
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
