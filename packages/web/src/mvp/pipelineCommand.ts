export function looksAbsoluteHostPath(p: string): boolean {
  const s = p.trim()
  if (!s) {
    return false
  }
  if (s.startsWith('/')) {
    return true
  }
  return /^[A-Za-z]:[\\/]/.test(s)
}
