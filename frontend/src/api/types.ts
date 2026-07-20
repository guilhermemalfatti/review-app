export type UserRole = 'resident' | 'admin' | string

export interface User {
  id: string
  email: string
  display_name: string
  role: UserRole
  condo_id: string
  must_change_password: boolean
}

export interface AdminUser {
  id: string
  email: string
  display_name: string
  role: UserRole
  must_change_password: boolean
  created_at: string
}

export interface ResetPasswordResult {
  user: AdminUser
  temporary_password: string
}

export interface ChangePasswordPayload {
  current_password: string
  new_password: string
}

export interface Aggregates {
  hired_count: number
  recommend_count: number
  not_recommend_count: number
  avg_price: number | null
  avg_quality: number | null
  avg_deadline: number | null
  avg_overall: number | null
  last_service_date: string | null
}

export interface ProviderListItem {
  id: string
  name: string
  category: string
  phone: string | null
  notes: string | null
  aggregates: Aggregates
}

export interface Review {
  id: string
  author_label: string
  recommend: boolean
  score_price: number | null
  score_quality: number | null
  score_deadline: number | null
  comment: string | null
  service_date: string | null
  created_at: string
}

export interface ProviderDetail extends ProviderListItem {
  reviews: Review[]
}

export interface PendingProvider {
  id: string
  name: string
  category: string
  phone: string | null
  notes: string | null
  status?: string
  created_at?: string
  suggested_by?: string
  creator_display_name?: string
  creator_email?: string
}

export interface PendingReview {
  id: string
  provider_id: string
  provider_name?: string
  author_label?: string
  author_display_name?: string
  author_email?: string
  recommend: boolean
  score_price: number | null
  score_quality: number | null
  score_deadline: number | null
  comment: string | null
  service_date: string | null
  created_at: string
  status?: string
}

export interface SignupPayload {
  email: string
  password: string
  display_name: string
  invite_code: string
}

export interface LoginPayload {
  email: string
  password: string
}

export interface CreateProviderPayload {
  name: string
  category: string
  phone: string
  notes: string
}

export interface CreateReviewPayload {
  is_anonymous: boolean
  recommend: boolean
  score_price?: number
  score_quality?: number
  score_deadline?: number
  comment: string
  service_date?: string
}

export interface MyReview {
  id: string
  is_anonymous: boolean
  recommend: boolean
  score_price: number | null
  score_quality: number | null
  score_deadline: number | null
  comment: string | null
  service_date: string | null
  status: string
  created_at: string
  updated_at: string
}

export const CATEGORIES = [
  'Eletricista',
  'Encanador',
  'Pintor',
  'Pedreiro',
  'Marceneiro',
  'Jardineiro',
  'Limpeza',
  'Ar-condicionado',
  'Outros',
] as const

export type Category = (typeof CATEGORIES)[number]
