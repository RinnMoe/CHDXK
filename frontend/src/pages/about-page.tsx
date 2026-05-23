import { PageShell } from "@/components/layout/page-shell"

export function AboutPage() {
  return (
    <>
      <title>关于 - JCourse</title>
      <PageShell>
        <div className="mx-auto max-w-3xl space-y-6">
          <div>
            <h1 className="text-2xl font-bold">关于 JCourse</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              JCourse 是面向学生的课程信息与课程点评平台，帮助用户查找课程、了解教师授课情况，并分享真实的学习体验。
            </p>
          </div>

          <section className="space-y-3 text-sm leading-7 text-muted-foreground">
            <p>
              我们希望通过结构化的课程信息、评分和点评，降低选课前的信息差，让每一次选课决策都有更多参考。
            </p>
            <p>
              平台内容由用户共同维护。发布点评时，请尽量提供具体、客观、可复核的信息，避免泄露个人隐私或发布无关内容。
            </p>
          </section>
        </div>
      </PageShell>
    </>
  )
}
