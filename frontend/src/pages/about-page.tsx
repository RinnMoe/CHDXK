import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { brand } from "@/config/brand"

export function AboutPage() {
  return (
    <>
      <PageTitle>关于</PageTitle>
      <PageShell>
        <div className="mx-auto max-w-3xl space-y-6">
          <div>
            <h1 className="text-2xl font-bold">关于 {brand.name}</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {brand.name}
              为非官方网站。选课社区目的在于让同学们了解课程的更多情况，不想也不能代替教务处的课程评教。
            </p>
          </div>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">机制</h2>
            <div className="space-y-2 text-sm leading-7 text-muted-foreground">
              <p>
                选课社区采用 jAccount
                或邮箱登录并作为身份标识。本站不明文存储您的 jAccount
                用户名，仅在数据库中存放其哈希值。
              </p>
              <p>
                选课社区前台不显示每条点评的用户名，也不显示不同点评之间的用户关联。
              </p>
            </div>
          </section>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">点评管理</h2>
            <div className="space-y-2 text-sm leading-7 text-muted-foreground">
              <p>
                在符合社区规范的情况下，我们不修改选课社区的点评内容，也不评价内容的真实性。如果您上过某一门课程并认为网站上的点评与事实不符，欢迎提交您的意见，我们相信全面的信息会给大家最好的答案。
              </p>
              <p>
                选课社区管理员的责任仅限于维护系统的稳定，删除非课程点评内容和重复发帖，并维护课程和教师信息格式，方便进行数据的批量处理。
              </p>
            </div>
          </section>

          <section className="space-y-3">
            <h2 className="text-lg font-semibold">隐私</h2>
            <div className="space-y-2 text-sm leading-7 text-muted-foreground">
              <p>
                当您访问选课社区时，我们使用百度统计收集您的访问信息，便于统计用户使用情况。
              </p>
              <p>
                当您登录选课社区时，我们会收集您的身份类型（在校生、教职工、校友等），但不收集除此以外的其他信息。
              </p>
              <p>
                选课社区部分功能可能需要使用 jAccount
                接口获取并存储选课等信息，我们将在您使用此类功能前予以提示。
              </p>
            </div>
          </section>

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
              联系我们。
            </p>
          </section>
        </div>
      </PageShell>
    </>
  )
}
