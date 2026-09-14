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
    title: "机制",
    paragraphs: [
      <>
        {brand.name} 使用部署方配置的统一身份服务登录。本站仅处理站点运行所需的账户映射与业务数据，课程记录来自部署方提供的课程数据。
      </>,
      <>{brand.name} 前台不展示用户的登录凭据；点评展示和账户关联以部署方实际规则为准。</>,
    ],
  },
  {
    title: "点评管理",
    paragraphs: [
      <>
        在符合社区规范的情况下，我们不修改 {brand.name} 的点评内容，也不评价内容的真实性。如果您上过某一门课程并认为网站上的点评与事实不符，欢迎提交您的意见，我们相信全面的信息会给大家最好的答案。
      </>,
      <>
        {brand.name} 管理员的责任仅限于维护系统的稳定，删除非课程点评内容和重复发帖，并维护课程和教师信息格式，方便进行数据的批量处理。
      </>,
    ],
  },
  {
    title: "隐私",
    paragraphs: [
      <>
        当您访问 {brand.name} 时，是否记录必要的访问日志或统计信息由部署配置决定，具体以部署方说明为准。
      </>,
      <>
        登录时，统一身份服务可能向本站提供完成登录所需的信息；本站仅保存运行所需的账户映射、操作记录和站点数据。
      </>,
      <>
        统一身份服务按照其隐私政策处理登录身份信息；具体数据范围和保存期限以当前部署方的说明为准。
      </>,
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
        <h1 className="text-2xl font-bold">关于 {brand.name}</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {brand.name}
          为非官方网站，旨在让同学们了解课程的更多情况，不想也不能代替教务处的课程评教。
        </p>
      </div>

      <AgreementSections />

      {brand.feedbackEmail && (
        <section className="space-y-3">
          <h2 className="text-lg font-semibold">联系方式</h2>
          <p className="text-sm leading-7 text-muted-foreground">
            您可以通过邮件{" "}
            <a
              href={`mailto:${brand.feedbackEmail}`}
              className="font-medium text-primary hover:underline"
            >
              {brand.feedbackEmail}
            </a>{" "}
            联系部署方。
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
        登录或注册即表示您了解并接受
        <DialogTrigger asChild>
          <Button variant="link" className="h-auto px-1 py-0 align-baseline">
            用户协议与站点说明
          </Button>
        </DialogTrigger>
        。
      </section>

      <DialogContent className="max-h-[calc(100dvh-4rem)] overflow-y-auto sm:max-w-3xl">
        <DialogTitle className="sr-only">用户协议与站点说明</DialogTitle>
        <DialogDescription className="sr-only">
          阅读 {brand.name} 的用户协议、站点说明和隐私说明。
        </DialogDescription>
        <AboutContent />
      </DialogContent>
    </Dialog>
  )
}
