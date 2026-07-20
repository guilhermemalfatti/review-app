import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { AuthProvider, useAuth } from './auth/AuthContext'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { AdminPage } from './pages/AdminPage'
import { ChangePasswordPage } from './pages/ChangePasswordPage'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { ProviderDetailPage } from './pages/ProviderDetailPage'
import { ReviewPage } from './pages/ReviewPage'
import { SignupPage } from './pages/SignupPage'
import { SuggestProviderPage } from './pages/SuggestProviderPage'

function PasswordGate({ children }: { children: React.ReactNode }) {
  const { user, loading, mustChangePassword } = useAuth()
  const location = useLocation()

  if (loading) return children
  if (user && mustChangePassword && location.pathname !== '/change-password') {
    return <Navigate to="/change-password" replace />
  }
  return children
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Layout>
          <PasswordGate>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/providers/new" element={
                <ProtectedRoute>
                  <SuggestProviderPage />
                </ProtectedRoute>
              } />
              <Route path="/providers/:id/review" element={
                <ProtectedRoute>
                  <ReviewPage />
                </ProtectedRoute>
              } />
              <Route path="/providers/:id" element={<ProviderDetailPage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/signup" element={<SignupPage />} />
              <Route path="/change-password" element={
                <ProtectedRoute allowPasswordChange>
                  <ChangePasswordPage />
                </ProtectedRoute>
              } />
              <Route path="/admin" element={
                <ProtectedRoute adminOnly>
                  <AdminPage />
                </ProtectedRoute>
              } />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </PasswordGate>
        </Layout>
      </AuthProvider>
    </BrowserRouter>
  )
}
