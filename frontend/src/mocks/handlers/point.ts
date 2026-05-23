import { http, HttpResponse } from "msw"
import type {
  CreatePointTransferCommand,
  PointTransferPreviewDTO,
} from "@/api/point"
import { findUserByID, mockSession, mockUsers } from "../fixtures/auth"
import {
  getMockPointRecords,
  getMockPointTotal,
  pushMockPointRecord,
} from "../fixtures/points"
import { randomDelay } from "../utils"

const FEE_RATE_BPS = 200
const MIN_FEE = 1

function calcFee(amount: number): number {
  return Math.max(Math.floor((amount * FEE_RATE_BPS) / 10000), MIN_FEE)
}

function calcPreview(
  cmd: CreatePointTransferCommand,
  senderBalance: number
): PointTransferPreviewDTO | { error: string; status: number } {
  const amount = Number(cmd.amount)
  if (!Number.isFinite(amount) || amount <= 0) {
    return { error: "invalid amount", status: 400 }
  }
  const feePayer = cmd.fee_payer === "recipient" ? "recipient" : "sender"
  const fee = calcFee(amount)
  if (feePayer === "recipient" && amount - fee <= 0) {
    return { error: "transfer amount too small", status: 400 }
  }
  return {
    amount,
    fee,
    fee_payer: feePayer,
    sender_debit: feePayer === "sender" ? amount + fee : amount,
    recipient_credit: feePayer === "recipient" ? amount - fee : amount,
    sender_balance: senderBalance,
    sender_remaining:
      senderBalance - (feePayer === "sender" ? amount + fee : amount),
  }
}

export const pointHandlers = [
  http.get("/api/user/:userID/points", async ({ params, request }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const userID = Number(params.userID)
    if (mockSession.userID !== userID) {
      const me = findUserByID(mockSession.userID)
      if (!me || me.role !== "admin") {
        return HttpResponse.json({ error: "forbidden" }, { status: 403 })
      }
    }
    const url = new URL(request.url)
    const page = Number(url.searchParams.get("page") ?? "1")
    const pageSize = Number(url.searchParams.get("page_size") ?? "20")

    const records = getMockPointRecords(userID)
    const start = (page - 1) * pageSize
    const items = records.slice(start, start + pageSize)
    return HttpResponse.json({
      total: getMockPointTotal(userID),
      records: {
        items,
        total: records.length,
        page,
        page_size: pageSize,
      },
    })
  }),

  http.post("/api/point/transfers/preview", async ({ request }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const cmd = (await request.json()) as CreatePointTransferCommand
    const preview = calcPreview(cmd, getMockPointTotal(mockSession.userID))
    if ("status" in preview) {
      return HttpResponse.json(
        { error: preview.error },
        { status: preview.status }
      )
    }
    return HttpResponse.json(preview)
  }),

  http.post("/api/point/transfers", async ({ request }) => {
    await randomDelay()
    if (!mockSession.userID) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const sender = findUserByID(mockSession.userID)
    if (!sender) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 })
    }
    const cmd = (await request.json()) as CreatePointTransferCommand
    const recipient = mockUsers.find(
      (u) => u.username === cmd.recipient_username
    )
    if (!recipient) {
      return HttpResponse.json(
        { error: "recipient not found" },
        { status: 404 }
      )
    }
    if (recipient.id === sender.id) {
      return HttpResponse.json(
        { error: "cannot transfer to self" },
        { status: 400 }
      )
    }
    const senderBalance = getMockPointTotal(sender.id)
    const preview = calcPreview(cmd, senderBalance)
    if ("status" in preview) {
      return HttpResponse.json(
        { error: preview.error },
        { status: preview.status }
      )
    }
    if (senderBalance < preview.sender_debit) {
      return HttpResponse.json(
        { error: "insufficient balance" },
        { status: 409 }
      )
    }

    const now = new Date().toISOString()
    pushMockPointRecord(sender.id, {
      reason: "transfer_out",
      amount: -preview.sender_debit,
      description: `转账给 ${recipient.email}`,
      created_at: now,
    })
    pushMockPointRecord(recipient.id, {
      reason: "transfer_in",
      amount: preview.recipient_credit,
      description: `收到 ${sender.email} 的转账`,
      created_at: now,
    })

    return HttpResponse.json(
      {
        id: Date.now(),
        sender_user_id: sender.id,
        recipient_user_id: recipient.id,
        amount: preview.amount,
        fee: preview.fee,
        fee_payer: preview.fee_payer,
        sender_delta: -preview.sender_debit,
        recipient_delta: preview.recipient_credit,
        created_at: now,
      },
      { status: 201 }
    )
  }),
]
