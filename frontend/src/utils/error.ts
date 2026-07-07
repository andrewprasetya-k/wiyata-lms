export function getApiError(error: unknown): string {
  if (typeof error === 'object' && error !== null) {
    const e = error as Record<string, unknown>
    if (e.response && typeof e.response === 'object') {
      const data = (e.response as Record<string, unknown>).data
      if (typeof data === 'string') return data
      if (typeof data === 'object' && data !== null) {
        const d = data as Record<string, unknown>
        if (typeof d.error === 'string') return d.error
        if (typeof d.message === 'string') return d.message
      }
    }
    if (typeof e.message === 'string') return e.message
  }
  return 'Terjadi kesalahan.'
}
