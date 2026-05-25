import type { PointRecordDTO } from "@/api/point"
import { mockUsers } from "./auth"

interface MockUserBalance {
  userID: number
  records: PointRecordDTO[]
}

const REASONS: {
  reason: PointRecordDTO["reason"]
  description: string
  amount: number
}[] = [
  { reason: "review_create", description: "发表点评奖励", amount: 5 },
  { reason: "review_create", description: "发表点评奖励", amount: 5 },
  { reason: "review_vote", description: "点评获得点赞", amount: 1 },
  { reason: "daily_login", description: "每日登录奖励", amount: 1 },
  {
    reason: "transfer_in",
    description: "积分转入",
    amount: 20,
  },
  {
    reason: "transfer_out",
    description: "积分转出",
    amount: -10,
  },
]

const balances = new Map<number, MockUserBalance>()

function ensureBalance(userID: number) {
  if (balances.has(userID)) return balances.get(userID)!
  const now = Date.now()
  const records: PointRecordDTO[] = REASONS.map((r, i) => ({
    reason: r.reason,
    description: r.description,
    amount: r.amount,
    created_at: new Date(now - i * 1000 * 60 * 60 * 24).toISOString(),
  }))
  const entry: MockUserBalance = { userID, records }
  balances.set(userID, entry)
  return entry
}

for (const u of mockUsers) {
  ensureBalance(u.id)
}

export function getMockPointRecords(userID: number): PointRecordDTO[] {
  return ensureBalance(userID).records
}

export function getMockPointTotal(userID: number): number {
  return getMockPointRecords(userID).reduce((sum, r) => sum + r.amount, 0)
}

export function pushMockPointRecord(userID: number, record: PointRecordDTO) {
  const entry = ensureBalance(userID)
  entry.records.unshift(record)
}
