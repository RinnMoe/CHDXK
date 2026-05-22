export interface ApiError {
  error: string
}

export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface RatingInfoDTO {
  count: number
  avg: number
  distribution: [number, number, number, number, number]
}

export interface TeacherDTO {
  id: number
  code: string
  name: string
  department: string
  title?: string
}

export interface FilterItem {
  name: string
  count: number
}
