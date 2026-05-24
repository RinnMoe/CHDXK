import type { ReviewDTO } from "@/api/review"
import { mockCourses } from "./courses"

const SCORES = ["A+", "A", "A-", "B+", "B", "B-", "C+", "C", "未公布"]
const SEMESTERS = ["2025-2026-1", "2024-2025-2", "2024-2025-1", "2023-2024-2"]

const SAMPLE_CONTENTS = [
  `讲得非常清楚，作业难度适中，考试也比较公平。

**推荐理由：**

- 老师备课充分，PPT 做得很认真
- 每节课都有 *随堂小测*，帮助巩固知识
- 期末考试范围明确，不超纲

> 总体来说是一门值得上的好课。`,

  `课程内容硬核，需要花大量时间预习和复习。老师水平很高，但对学生要求也高。

## 课程结构

1. 前半学期：理论基础
2. 后半学期：项目实践

**作业量：** 每周约 4-6 小时，建议组队完成。

评分构成：

| 项目 | 占比 |
|------|------|
| 平时作业 | 30% |
| 期中考试 | 20% |
| 期末项目 | 50% |`,

  `比较水的课，平时点名加期末小论文，不挂人。适合刷绩点。

适合以下同学：

- 想轻松拿学分的
- 对该领域只想了解个大概的
- 时间紧张需要平衡其他硬课的`,

  `推荐！老师人很好，讲课节奏适中，PPT 也做得很认真。

课程亮点：

1. **课堂互动多** — 每节课都有讨论环节
2. *案例丰富* — 结合实际工程问题讲解
3. \`代码示例\` 都可以直接运行

不过有一点要注意：**期中考试比较难**，一定要提前复习。`,

  `课程难度偏高，但内容很有意思。注意期中考试占 40%，平时一定要跟上。

### 学习建议

- **预习**：至少提前看一遍教材对应章节
- **复习**：课后整理笔记，重点理解公式推导
- **刷题**：往年题很有参考价值

> 这门课虽然辛苦，但学完之后收获很大，对后续课程帮助也多。`,

  `授课方式比较传统，主要是 PPT 念读。建议自学为主。

**自学资源推荐：**

- [MIT OCW](https://ocw.mit.edu) 对应课程视频
- 教材课后习题（考试很多原题改编）
- B 站上也有不错的中文讲解

如果只是想拿学分，考前突击一周基本够了。`,

  `老师讲得很有激情，但内容太多，每节课信息量爆炸。考试前需要好好整理笔记。

## 优缺点

**优点：**
- 老师知识面广，能回答各种延伸问题
- 办公时间（office hour）很耐心

**缺点：**
- 进度太快，稍不留神就听不懂
- 作业偏难，deadline 卡得紧`,

  `总体不错，作业有些多但都很有针对性。期末考试题目不算难，但需要灵活应用。

### 评分

- 课程难度：⭐⭐⭐⭐
- 老师教学：⭐⭐⭐⭐⭐
- 获益程度：⭐⭐⭐⭐

一句话总结：**投入和回报成正比的一门课。**`,
]

function randInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function pick<T>(arr: T[]): T {
  return arr[randInt(0, arr.length - 1)]
}

let reviewIDSeed = 1

export function generateReviewsForCourse(courseID: number, n = 8): ReviewDTO[] {
  const course = mockCourses.find((c) => c.id === courseID)
  return Array.from({ length: n }, (_, i) => {
    const created = new Date(Date.now() - randInt(1, 365) * 86400000)
    // ~30% chance the review was edited later
    const updated =
      randInt(0, 9) < 3
        ? new Date(
            created.getTime() +
              randInt(1, 30) * 86400000 +
              randInt(0, 86400) * 1000
          )
        : created
    return {
      id: reviewIDSeed++,
      course_id: courseID,
      course,
      user_id: i < 2 ? 1 : 0,
      semester: pick(SEMESTERS),
      score: pick(SCORES),
      rating: randInt(1, 5),
      content: pick(SAMPLE_CONTENTS),
      moderator_remark:
        randInt(0, 9) < 2 ? pick(["内容已核实", "请注意描述规范", ""]) : "",
      vote: {
        like_count: randInt(0, 30),
        dislike_count: randInt(0, 5),
      },
      created_at: created.toISOString(),
      updated_at: updated.toISOString(),
    } as ReviewDTO
  })
}

export const mockReviews: ReviewDTO[] = mockCourses.flatMap((c) =>
  generateReviewsForCourse(c.id, randInt(18, 28))
)

mockReviews[0] = {
  ...mockReviews[0],
  rating: 5,
  content: `这条 mock 用来检查点评引用链接渲染效果：#2 和 #3 应该会变成可点击的点评链接。

普通 Markdown 链接仍保持原样：[课程详情](/course/${mockReviews[0].course_id})。

代码里的引用不应该被改写：

\`#4\`

\`\`\`txt
#5
\`\`\`

相邻文字里的格式也能识别，比如“参考 #6 的补充”。`,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

export function findReview(id: number): ReviewDTO | undefined {
  return mockReviews.find((r) => r.id === id)
}
