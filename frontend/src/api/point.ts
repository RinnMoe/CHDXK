import type { PaginatedResult } from "./types"
import { apiClient } from "./client"
import { BASE_URL } from "./constants"

export type FeePayer = "sender" | "recipient"

export type RecordReason = "transfer_out" | "transfer_in" | string

export interface PointRecordDTO {
  reason: RecordReason
  amount: number
  description: string
  created_at: string
}

export interface PointSummaryDTO {
  total: number
  records: PaginatedResult<PointRecordDTO>
}

export interface PointTransferDTO {
  id: number
  sender_user_id: number
  recipient_user_id: number
  amount: number
  fee: number
  fee_payer: FeePayer
  sender_delta: number
  recipient_delta: number
  created_at: string
}

export interface PointTransferPreviewDTO {
  amount: number
  fee: number
  fee_payer: FeePayer
  sender_debit: number
  recipient_credit: number
  sender_balance: number
  sender_remaining: number
}

export interface CreatePointTransferCommand {
  recipient_email: string
  amount: number
  fee_payer: FeePayer
}

export interface PreviewTransferParams {
  amount: number
  fee_payer: FeePayer
}

export interface PointRecordListFilter {
  page?: number
  page_size?: number
  [key: string]: unknown
}

function buildQuery(filter: Record<string, unknown>): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filter)) {
    if (value === undefined || value === null || value === "") continue
    params.append(key, String(value))
  }
  const q = params.toString()
  return q ? `?${q}` : ""
}

export function getUserPoints(
  userID: number,
  filter: PointRecordListFilter = {}
): Promise<PointSummaryDTO> {
  return apiClient(`${BASE_URL}/user/${userID}/point${buildQuery(filter)}`)
}

export function previewTransfer(
  params: PreviewTransferParams
): Promise<PointTransferPreviewDTO> {
  return apiClient(`${BASE_URL}/point/transfer/preview`, {
    method: "POST",
    body: JSON.stringify(params),
  })
}

export function createTransfer(
  cmd: CreatePointTransferCommand
): Promise<PointTransferDTO> {
  return apiClient(`${BASE_URL}/point/transfer`, {
    method: "POST",
    body: JSON.stringify(cmd),
  })
}
