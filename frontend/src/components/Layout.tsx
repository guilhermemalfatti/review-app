import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { COMMUNITY_NAME } from '../config'

export function Layout({ children }: { children: React.ReactNode }) {
  const { user, loading, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/')
  }

  return (
    <div className="app-shell">
      <a href="#main-content" className="skip-link">
        Ir para o conteúdo
      </a>

      <header className="site-header">
        <div className="site-header__inner">
          <Link to="/" className="brand" aria-label="Indica — início">
            <span className="brand__mark" aria-hidden="true" />
            <span className="brand__name">Indica {COMMUNITY_NAME}</span>
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
                Aprovar
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

      <main id="main-content" className="site-main" tabIndex={-1}>
        {children}
      </main>

      <footer className="site-footer">
        <p>
          Indica · recomendações entre vizinhos · <span>{COMMUNITY_NAME}</span>
        </p>
      </footer>
    </div>
  )
}
