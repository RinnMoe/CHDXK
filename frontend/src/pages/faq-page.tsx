import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"

const faqs = [
  {
    question: "如何查找课程？",
    answer:
      "可以在课程页通过关键词、院系、分类、学分等条件筛选课程，也可以直接查看热门课程。",
  },
  {
    question: "如何发布课程点评？",
    answer:
      "登录后进入课程详情页，点击发布点评即可提交对课程和教师授课体验的评价。",
  },
  {
    question: "点评内容可以修改吗？",
    answer:
      "已发布的点评可以在个人点评列表或点评详情页进入编辑流程，按实际体验更新内容。",
  },
  {
    question: "为什么部分课程信息不完整？",
    answer:
      "课程数据可能来自不同批次的导入与用户补充。如果发现信息缺失或不准确，可以通过点评补充上下文。",
  },
]

export function FaqPage() {
  return (
    <>
      <PageTitle>常见问题</PageTitle>
      <PageShell>
        <div className="mx-auto max-w-3xl space-y-6">
          <div>
            <h1 className="text-2xl font-bold">常见问题</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              关于课程查询、点评发布和内容维护的常见说明。
            </p>
          </div>

          <div className="divide-y rounded-lg border bg-card">
            {faqs.map((faq) => (
              <section key={faq.question} className="space-y-2 p-4">
                <h2 className="font-medium">{faq.question}</h2>
                <p className="text-sm leading-7 text-muted-foreground">
                  {faq.answer}
                </p>
              </section>
            ))}
          </div>
        </div>
      </PageShell>
    </>
  )
}
