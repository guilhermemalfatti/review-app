import { useState } from 'react'
import { Link } from 'react-router-dom'
import { COMMUNITY_NAME, PIX_KEY, PIX_KEY_LABEL } from '../config'

const SUGGESTED_AMOUNTS = [10, 20, 30, 40, 50] as const

export function DonatePage() {
  const [copied, setCopied] = useState(false)
  const [copyError, setCopyError] = useState(false)
  const hasPixKey = PIX_KEY.trim().length > 0

  async function handleCopy() {
    if (!hasPixKey) return
    setCopyError(false)
    try {
      await navigator.clipboard.writeText(PIX_KEY.trim())
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2500)
    } catch {
      setCopyError(true)
    }
  }

  return (
    <div className="page page--donate page-enter">
      <Link to="/" className="text-link text-link--back">
        ← Prestadores
      </Link>

      <header className="donate-header">
        <p className="donate-header__eyebrow">Apoie o Indica</p>
        <h1 className="donate-header__title">Um café para manter a comunidade</h1>
        <p className="donate-header__lead">
          O Indica é gratuito para os vizinhos do {COMMUNITY_NAME}. Se o app te ajudou a
          achar um prestador de confiança, um PIX simbólico ajuda a pagar hospedagem e
          seguir melhorando.
        </p>
      </header>

      <section className="donate-pix" aria-labelledby="pix-heading">
        <h2 id="pix-heading" className="donate-pix__title">
          Enviar um PIX
        </h2>
        <p className="donate-pix__hint">
          Abra o app do seu banco, escolha PIX, cole a chave e digite o valor que quiser.
        </p>

        {hasPixKey ? (
          <>
            <div className="donate-pix__key-block">
              <span className="donate-pix__key-label">{PIX_KEY_LABEL}</span>
              <code className="donate-pix__key">{PIX_KEY.trim()}</code>
            </div>
            <button
              type="button"
              className="btn btn--primary btn--block"
              onClick={() => void handleCopy()}
            >
              {copied ? 'Chave copiada!' : 'Copiar chave PIX'}
            </button>
            {copyError && (
              <p className="donate-pix__error" role="alert">
                Não foi possível copiar. Selecione a chave e copie manualmente.
              </p>
            )}
          </>
        ) : (
          <p className="donate-pix__pending">
            A chave PIX ainda não foi configurada. Volte em breve.
          </p>
        )}
      </section>

      <section className="donate-amounts" aria-labelledby="amounts-heading">
        <h2 id="amounts-heading" className="donate-amounts__title">
          Sugestões de valor
        </h2>
        <p className="donate-amounts__hint">Qualquer valor ajuda — estas são só ideias.</p>
        <ul className="donate-amounts__list">
          {SUGGESTED_AMOUNTS.map((amount) => (
            <li key={amount} className="donate-amounts__item">
              R$ {amount}
            </li>
          ))}
        </ul>
      </section>

      <p className="donate-thanks">Obrigado por apoiar o Indica.</p>
    </div>
  )
}
