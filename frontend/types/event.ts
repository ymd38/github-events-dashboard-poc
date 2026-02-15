export interface Event {
  id: number
  delivery_id: string
  event_type: string
  action: string
  repo_name: string
  sender_login: string
  sender_avatar_url: string | null
  title: string | null
  body: string | null
  html_url: string
  event_data: Record<string, unknown> | null
  occurred_at: string
  received_at: string
}

export interface Pagination {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface EventListResponse {
  events: Event[]
  pagination: Pagination
}
