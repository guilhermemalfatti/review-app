import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'

export function Layout({ children }: { children: React.ReactNode }) {
  const { user, loading, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/')
  }

  return (
    <div className="app-shell">
      <header className="site-header">
        <div className="site-header__inner">
          <Link to="/" className="brand" aria-label="Indica — início">
            <span className="brand__mark" aria-hidden="true" />
            <span className="brand__name">Indica</span>
          </Link>

          <nav className="site-nav" aria-label="Principal">
            <NavLink to="/" end className={({ isActive }) => (isActive ? 'is-active' : undefined)}>
              Prestadores
            </NavLink>
            <NavLink
              to="/providers/new"
              className={({ isActive }) => (isActive ? 'is-active' : undefined)}
            >
              Sugerir
            </NavLink>
            {isAdmin && (
              <NavLink
                to="/admin"
                className={({ isActive }) => (isActive ? 'is-active' : undefined)}
              >
                Admin
              </NavLink>
            )}
            {!loading &&
              (user ? (
                <button type="button" className="nav-button" onClick={() => void handleLogout()}>
                  Sair
                </button>
              ) : (
                <NavLink
                  to="/login"
                  className={({ isActive }) => (isActive ? 'is-active' : undefined)}
                >
                  Entrar
                </NavLink>
              ))}
          </nav>
        </div>
      </header>

      <main className="site-main">{children}</main>

      <footer className="site-footer">
        <p>
          Indica · recomendações entre vizinhos · <span>Cantegril</span>
        </p>
      </footer>
    </div>
  )
}
