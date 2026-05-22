import type {
  CourseListItemDTO,
  CourseDetailDTO,
  CourseFilters,
} from "@/api/course"

const DEPARTMENTS = [
  "计算机科学与工程系",
  "数学科学学院",
  "物理与天文学院",
  "电子信息与电气工程学院",
  "外国语学院",
  "化学化工学院",
  "生命科学技术学院",
  "机械与动力工程学院",
]

const CATEGORIES = ["专业必修", "专业选修", "通识核心", "通识选修", "公共基础"]

const TARGET_YEARS = ["大一", "大二", "大三", "大四", "研究生"]

const LANGUAGES = ["中文", "英文", "双语"]

const CREDITS = [1, 2, 3, 4]

const COURSE_NAMES = [
  "数据结构",
  "操作系统",
  "计算机网络",
  "编译原理",
  "高等数学",
  "线性代数",
  "概率论与数理统计",
  "离散数学",
  "大学物理",
  "理论力学",
  "电磁学",
  "量子力学",
  "大学英语",
  "学术英语写作",
  "有机化学",
  "无机化学",
  "细胞生物学",
  "分子生物学",
  "机械设计基础",
  "热力学",
  "人工智能导论",
  "机器学习",
  "深度学习",
  "数据库系统",
  "软件工程",
  "算法分析与设计",
  "计算机组成原理",
  "信号与系统",
  "数字电路",
  "模拟电路",
  "自然语言处理",
  "计算机图形学",
  "并行计算",
  "分布式系统",
  "网络安全",
  "区块链技术",
  "云计算",
  "嵌入式系统",
  "数据挖掘",
  "信息检索",
]

const TEACHER_NAMES = [
  "张伟",
  "王芳",
  "李娜",
  "刘洋",
  "陈静",
  "杨磊",
  "黄丽",
  "赵勇",
  "周敏",
  "吴杰",
  "徐峰",
  "孙颖",
  "马超",
  "朱琳",
  "胡军",
  "郭强",
  "何萍",
]

function randInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function pick<T>(arr: T[]): T {
  return arr[randInt(0, arr.length - 1)]
}

function pickMany<T>(arr: T[], n: number): T[] {
  const copy = [...arr]
  const result: T[] = []
  for (let i = 0; i < n && copy.length > 0; i++) {
    const idx = randInt(0, copy.length - 1)
    result.push(copy.splice(idx, 1)[0])
  }
  return result
}

let teacherIDSeed = 1
const teacherCache = new Map<string, { id: number; code: string; name: string; department: string; title?: string }>()

function makeTeacher(name?: string) {
  const finalName = name ?? pick(TEACHER_NAMES)
  if (teacherCache.has(finalName)) return teacherCache.get(finalName)!
  const dept = pick(DEPARTMENTS)
  const teacher = {
    id: teacherIDSeed++,
    code: `T${String(teacherIDSeed).padStart(5, "0")}`,
    name: finalName,
    department: dept,
    title: pick(["教授", "副教授", "讲师", "助理教授"]),
  }
  teacherCache.set(finalName, teacher)
  return teacher
}

function makeDistribution(count: number): [number, number, number, number, number] {
  if (count === 0) return [0, 0, 0, 0, 0]
  const dist: [number, number, number, number, number] = [0, 0, 0, 0, 0]
  for (let i = 0; i < count; i++) {
    dist[randInt(0, 4)]++
  }
  return dist
}

function avgFromDistribution(dist: [number, number, number, number, number]): number {
  const total = dist.reduce((a, b) => a + b, 0)
  if (total === 0) return 0
  const sum = dist.reduce((s, c, i) => s + c * (i + 1), 0)
  return Math.round((sum / total) * 100) / 100
}

function makeCourse(id: number): CourseListItemDTO {
  const count = randInt(0, 50)
  const distribution = makeDistribution(count)
  const mainTeacher = makeTeacher()
  return {
    id,
    code: `CS${String(id).padStart(4, "0")}`,
    name: pick(COURSE_NAMES),
    credit: pick(CREDITS),
    department: mainTeacher.department,
    language: pick(LANGUAGES),
    target_years: pickMany(TARGET_YEARS, randInt(1, 2)),
    categories: pickMany(CATEGORIES, randInt(1, 2)),
    main_teacher: mainTeacher,
    rating: {
      count,
      avg: avgFromDistribution(distribution),
      distribution,
    },
  }
}

export const mockCourses: CourseListItemDTO[] = Array.from({ length: 64 }, (_, i) =>
  makeCourse(i + 1)
)

mockCourses[0] = {
  ...mockCourses[0],
  name: "面向复杂真实世界系统的超大规模分布式数据库架构设计与性能调优实践",
  department: "电子信息与电气工程学院",
  language: "中文及英文双语研讨",
  categories: ["跨学科综合实践课程", "研究型专业选修模块"],
  main_teacher: makeTeacher("欧阳明远清和"),
}

mockCourses[1] = {
  ...mockCourses[1],
  name: "人工智能安全、可信机器学习与大模型治理专题前沿导论",
  department: "计算机科学与工程系",
  categories: ["通识核心-科技伦理与社会", "专业方向拓展"],
  main_teacher: makeTeacher("司徒嘉言"),
}

mockCourses[2] = {
  ...mockCourses[2],
  name: "计算社会科学中的因果推断、网络实验与高维数据分析方法",
  department: "数学科学学院",
  language: "全英文授课与中文讨论",
  categories: ["方法论强化训练", "通识选修-社会科学"],
  main_teacher: makeTeacher("Alexander Christopher Johnson-Smith"),
}

export function makeCourseDetail(course: CourseListItemDTO): CourseDetailDTO {
  const sameCode = mockCourses
    .filter((c) => c.id !== course.id)
    .slice(0, 3)
  const sameTeacher = mockCourses
    .filter((c) => c.main_teacher.id === course.main_teacher.id && c.id !== course.id)
    .slice(0, 3)

  return {
    id: course.id,
    code: course.code,
    name: course.name,
    credit: course.credit,
    department: course.department,
    language: course.language,
    target_years: course.target_years,
    categories: course.categories,
    main_teacher: course.main_teacher,
    rating: course.rating,
    notification_level: 0,
    offered_courses: [
      {
        semester: "2025-2026-1",
        language: course.language,
        target_years: course.target_years,
        categories: course.categories,
        teacher_group: [
          course.main_teacher,
          makeTeacher(pick(TEACHER_NAMES)),
        ],
      },
      {
        semester: "2024-2025-2",
        language: course.language,
        target_years: course.target_years,
        categories: course.categories,
        teacher_group: [course.main_teacher],
      },
    ],
    same_code_courses: sameCode,
    same_teacher_courses: sameTeacher,
  }
}

export function makeCourseFilters(): CourseFilters {
  const counts = (arr: string[]) =>
    arr.map((name) => ({
      name,
      count: randInt(5, 60),
    }))
  return {
    credits: CREDITS.map((c) => ({ name: String(c), count: randInt(5, 30) })),
    departments: counts(DEPARTMENTS),
    categories: counts(CATEGORIES),
    target_years: counts(TARGET_YEARS),
  }
}
