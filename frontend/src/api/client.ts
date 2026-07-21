import type {
  AdminUser,
  ChangePasswordPayload,
  CreatedProvider,
  CreateProviderPayload,
  CreateReviewPayload,
  LoginPayload,
  MyReview,
  PendingProvider,
  PendingReview,
  ProviderDetail,
  ProviderListItem,
  ResetPasswordResult,
  SignupPayload,
  User,
} from './types'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

let csrfToken: string | null = null

/** Empty in local dev (Vite proxies `/api`). Absolute origin on GitHub Pages. */
const API_BASE = String(import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '')

function apiUrl(path: string): string {
  return `${API_BASE}${path}`
}

async function parseError(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as { error?: string; message?: string }
    return data.error || data.message || `Erro ${response.status}`
  } catch {
    return `Erro ${response.status}`
  }
}

function isMutatingMethod(method: string): boolean {
  return method === 'POST' || method === 'PUT' || method === 'PATCH' || method === 'DELETE'
}

function isCsrfRelated(message: string): boolean {
  return /csrf/i.test(message)
}

async function ensureCsrf(): Promise<string> {
  if (csrfToken) return csrfToken

  const response = await fetch(apiUrl('/api/auth/csrf'), {
    credentials: 'include',
  })

  if (!response.ok) {
    throw new ApiError(response.status, await parseError(response))
  }

  const data = (await response.json()) as { csrf_token: string }
  csrfToken = data.csrf_token
  return csrfToken
}

async function request<T>(path: string, init?: RequestInit, retried = false): Promise<T> {
  const method = (init?.method ?? 'GET').toUpperCase()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (init?.headers) {
    const incoming = new Headers(init.headers)
    incoming.forEach((value, key) => {
      headers[key] = value
    })
  }

  if (isMutatingMethod(method)) {
    const token = await ensureCsrf()
    headers['X-CSRF-Token'] = token
  }

  const response = await fetch(apiUrl(path), {
    ...init,
    method,
    credentials: 'include',
    headers,
  })

  if (!response.ok) {
    const message = await parseError(response)
    if (
      response.status === 403 &&
      isCsrfRelated(message) &&
      isMutatingMethod(method) &&
      !retried
    ) {
      csrfToken = null
      await ensureCsrf()
      return request<T>(path, init, true)
    }
    throw new ApiError(response.status, message)
  }

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  if (!text) {
    return undefined as T
  }

  return JSON.parse(text) as T
}

export const api = {
  signup(payload: SignupPayload) {
    return request<{ user: User }>('/api/auth/signup', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  login(payload: LoginPayload) {
    return request<{ user: User }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  logout() {
    return request<void>('/api/auth/logout', { method: 'POST' })
  },

  me() {
    return request<{ user: User }>('/api/auth/me')
  },

  changePassword(payload: ChangePasswordPayload) {
    return request<{ user: User }>('/api/auth/change-password', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  listProviders(params?: { category?: string; q?: string }) {
    const search = new URLSearchParams()
    if (params?.category) search.set('category', params.category)
    if (params?.q) search.set('q', params.q)
    const query = search.toString()
    return request<ProviderListItem[]>(`/api/providers${query ? `?${query}` : ''}`)
  },

  getProvider(id: string) {
    return request<ProviderDetail>(`/api/providers/${id}`)
  },

  createProvider(payload: CreateProviderPayload) {
    return request<CreatedProvider>('/api/providers', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  createReview(providerId: string, payload: CreateReviewPayload) {
    return request<unknown>(`/api/providers/${providerId}/reviews`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  getMyReview(providerId: string) {
    return request<MyReview>(`/api/providers/${providerId}/my-review`)
  },

  adminPendingProviders() {
    return request<PendingProvider[]>('/api/admin/providers?status=pending')
  },

  adminApprovedProviders() {
    return request<PendingProvider[]>('/api/admin/providers?status=approved')
  },

  adminApproveProvider(id: string) {
    return request<void>(`/api/admin/providers/${id}/approve`, { method: 'POST' })
  },

  adminRejectProvider(id: string) {
    return request<void>(`/api/admin/providers/${id}/reject`, { method: 'POST' })
  },

  adminRemoveProvider(id: string) {
    return request<void>(`/api/admin/providers/${id}/remove`, { method: 'POST' })
  },

  adminPendingReviews() {
    return request<PendingReview[]>('/api/admin/reviews?status=pending')
  },

  adminApproveReview(id: string) {
    return request<void>(`/api/admin/reviews/${id}/approve`, { method: 'POST' })
  },

  adminRejectReview(id: string) {
    return request<void>(`/api/admin/reviews/${id}/reject`, { method: 'POST' })
  },

  adminListUsers() {
    return request<AdminUser[]>('/api/admin/users')
  },

  adminResetPassword(id: string) {
    return request<ResetPasswordResult>(`/api/admin/users/${id}/reset-password`, {
      method: 'POST',
    })
  },
}
