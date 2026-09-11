import { useEffect, useRef, useState } from 'react'
import { Link, NavLink, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { COMMUNITY_NAME } from '../config'

export function Layout({ children }: { children: React.ReactNode }) {
  const { user, loading, isAdmin, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [menuOpen, setMenuOpen] = useState(false)
  const toggleRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLElement>(null)
  const wasOpenRef = useRef(false)

  useEffect(() => {
    setMenuOpen(false)
  }, [location.pathname])

  useEffect(() => {
    if (!menuOpen) return

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setMenuOpen(false)
      }
    }

    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [menuOpen])

  useEffect(() => {
    if (menuOpen) {
      wasOpenRef.current = true
      const firstFocusable = menuRef.current?.querySelector<HTMLElement>(
        'a, button:not([disabled])',
      )
      firstFocusable?.focus()
      return
    }

    if (wasOpenRef.current) {
      wasOpenRef.current = false
      toggleRef.current?.focus()
    }
  }, [menuOpen])

  async function handleLogout() {
    setMenuOpen(false)
    await logout()
    navigate('/')
  }

  function closeMenu() {
    setMenuOpen(false)
  }

  return (
    <div className="app-shell">
      <a href="#main-content" className="skip-link">
        Ir para o conteúdo
      </a>

      <header className={`site-header${menuOpen ? ' site-header--menu-open' : ''}`}>
        <div className="site-header__inner">
          <Link to="/" className="brand" aria-label="Indica — início">
            <span className="brand__mark" aria-hidden="true" />
            <span className="brand__name">Indica {COMMUNITY_NAME}</span>
          </Link>

          <button
            ref={toggleRef}
            type="button"
            className={`site-menu-toggle${menuOpen ? ' is-open' : ''}`}
            aria-expanded={menuOpen}
            aria-controls="site-menu"
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span className="site-menu-toggle__icon" aria-hidden="true" />
            <span className="sr-only">{menuOpen ? 'Fechar menu' : 'Menu'}</span>
          </button>

          <nav
            ref={menuRef}
            id="site-menu"
            className={`site-nav${menuOpen ? ' is-open' : ''}`}
            aria-label="Principal"
          >
            <NavLink
              to="/"
              end
              className={({ isActive }) => (isActive ? 'is-active' : undefined)}
              onClick={closeMenu}
            >
              Prestadores
            </NavLink>
            <NavLink
              to="/apoiar"
              className={({ isActive }) => (isActive ? 'is-active' : undefined)}
              onClick={closeMenu}
            >
              Apoiar
            </NavLink>
            <NavLink
              to="/sobre"
              className={({ isActive }) => (isActive ? 'is-active' : undefined)}
              onClick={closeMenu}
            >
              Sobre
            </NavLink>
            <NavLink
              to="/providers/new"
              className={({ isActive }) => (isActive ? 'is-active' : undefined)}
              onClick={closeMenu}
            >
              Sugerir
            </NavLink>
            {isAdmin && (
              <NavLink
                to="/admin"
                className={({ isActive }) => (isActive ? 'is-active' : undefined)}
                onClick={closeMenu}
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
                  <button
                    type="button"
                    className="nav-button"
                    onClick={() => void handleLogout()}
                  >
                    Sair
                  </button>
                </>
              ) : (
                <NavLink
                  to="/login"
                  className={({ isActive }) => (isActive ? 'is-active' : undefined)}
                  onClick={closeMenu}
                >
                  Entrar
                </NavLink>
              ))}
          </nav>
        </div>

        {menuOpen && (
          <button
            type="button"
            className="site-menu-backdrop"
            aria-label="Fechar menu"
            onClick={closeMenu}
          />
        )}
      </header>

      <main id="main-content" className="site-main" tabIndex={-1}>
        {children}
      </main>

      <footer className="site-footer">
        <p>
          Indica · recomendações entre vizinhos · <span>{COMMUNITY_NAME}</span>
        </p>
        <p className="site-footer__support">
          <Link to="/sobre">Sobre</Link>
          {' · '}
          <Link to="/apoiar">Apoie com um café</Link>
        </p>
      </footer>
    </div>
  )
}
