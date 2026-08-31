import { Link } from 'react-router-dom'
import { COMMUNITY_NAME } from '../config'
import { whatsAppURL } from '../lib/phone'
import { usePageTitle } from '../lib/usePageTitle'

const ADMINS = [
  {
    name: 'Gilmara',
    phone: '+55 51 99956-3242',
  },
] as const

export function AboutPage() {
  usePageTitle('Sobre · Indica')

  return (
    <div className="page page--about page-enter">
      <Link to="/" className="text-link text-link--back">
        ← Prestadores
      </Link>

      <header className="about-header">
        <p className="about-header__eyebrow">Sobre o Indica</p>
        <h1 className="about-header__title">Indica {COMMUNITY_NAME}</h1>
        <p className="about-header__lead">
          O Indica reúne indicações de prestadores entre vizinhos. Se precisar de ajuda
          com a conta — por exemplo, redefinir a senha — fale com um administrador.
        </p>
      </header>

      <section className="about-admins" aria-labelledby="admins-heading">
        <h2 id="admins-heading" className="about-admins__title">
          Administradores
        </h2>
        <ul className="about-admins__list">
          {ADMINS.map((admin) => {
            const wa = whatsAppURL(admin.phone)
            return (
              <li key={admin.name} className="about-admins__item">
                <p className="about-admins__name">{admin.name}</p>
                <p className="about-admins__phone">
                  <a href={`tel:${admin.phone.replace(/\D/g, '')}`}>{admin.phone}</a>
                </p>
                {wa && (
                  <a
                    href={wa}
                    className="btn btn--ghost"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    WhatsApp
                  </a>
                )}
              </li>
            )
          })}
        </ul>
      </section>
    </div>
  )
}
