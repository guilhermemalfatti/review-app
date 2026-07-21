import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, MustChangePasswordGate } from './auth/AuthContext'
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

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Layout>
          <MustChangePasswordGate>
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
                <ProtectedRoute>
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
          </MustChangePasswordGate>
        </Layout>
      </AuthProvider>
    </BrowserRouter>
  )
}
