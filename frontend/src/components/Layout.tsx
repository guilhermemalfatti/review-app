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
              to="/apoiar"
              className={({ isActive }) => (isActive ? 'is-active' : undefined)}
            >
              Apoiar
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
                <>
                  <span className="site-nav__user" title={user.email}>
                    Olá, {user.display_name}
                  </span>
                  <button type="button" className="nav-button" onClick={() => void handleLogout()}>
                    Sair
                  </button>
                </>
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
        <p className="site-footer__support">
          <Link to="/apoiar">Apoie com um café</Link>
        </p>
      </footer>
    </div>
  )
}
