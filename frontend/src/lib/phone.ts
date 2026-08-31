/** Digits only for tel: / wa.me links. */
export function formatPhoneDigits(phone: string): string {
  return phone.replace(/\D/g, '')
}

/** Brazilian mobile → wa.me URL. Assumes local numbers without country code get 55. */
export function whatsAppURL(phone: string): string | null {
  const digits = formatPhoneDigits(phone)
  if (digits.length < 10) return null
  const withCountry = digits.startsWith('55') ? digits : `55${digits}`
  return `https://wa.me/${withCountry}`
}
