import type {
  CreateProviderPayload,
  CreateReviewPayload,
  LoginPayload,
  PendingProvider,
  PendingReview,
  ProviderDetail,
  ProviderListItem,
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

async function parseError(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as { error?: string; message?: string }
    return data.error || data.message || `Erro ${response.status}`
  } catch {
    return `Erro ${response.status}`
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  if (!response.ok) {
    throw new ApiError(response.status, await parseError(response))
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

  listProviders(params?: { category?: string; q?: string }) {
    const search = new URLSearchParams()
    if (params?.category) search.set('category', params.category)
    if (params?.q) search.set('q', params.q)
    const query = search.toString()
    return request<ProviderListItem[] | { providers: ProviderListItem[] }>(
      `/api/providers${query ? `?${query}` : ''}`,
    )
  },

  getProvider(id: string) {
    return request<ProviderDetail>(`/api/providers/${id}`)
  },

  createProvider(payload: CreateProviderPayload) {
    return request<ProviderListItem>('/api/providers', {
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

  adminPendingProviders() {
    return request<PendingProvider[] | { providers: PendingProvider[] }>(
      '/api/admin/providers?status=pending',
    )
  },

  adminApproveProvider(id: string) {
    return request<void>(`/api/admin/providers/${id}/approve`, { method: 'POST' })
  },

  adminRejectProvider(id: string) {
    return request<void>(`/api/admin/providers/${id}/reject`, { method: 'POST' })
  },

  adminPendingReviews() {
    return request<PendingReview[] | { reviews: PendingReview[] }>(
      '/api/admin/reviews?status=pending',
    )
  },

  adminApproveReview(id: string) {
    return request<void>(`/api/admin/reviews/${id}/approve`, { method: 'POST' })
  },

  adminRejectReview(id: string) {
    return request<void>(`/api/admin/reviews/${id}/reject`, { method: 'POST' })
  },
}

export function unwrapList<T>(data: T[] | { providers: T[] } | { reviews: T[] }): T[] {
  if (Array.isArray(data)) return data
  if ('providers' in data) return data.providers
  if ('reviews' in data) return data.reviews
  return []
}
