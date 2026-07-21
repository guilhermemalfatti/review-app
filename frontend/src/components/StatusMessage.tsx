interface StatusMessageProps {
  tone?: 'info' | 'success' | 'error'
  children: React.ReactNode
}

export function StatusMessage({ tone = 'info', children }: StatusMessageProps) {
  return (
    <div
      className={`status-message status-message--${tone}`}
      role={tone === 'error' ? 'alert' : 'status'}
    >
      {children}
    </div>
  )
}
