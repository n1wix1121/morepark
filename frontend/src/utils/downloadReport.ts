export async function downloadExcelReport(endpoint: string, fallbackFilename: string): Promise<void> {
  const token = localStorage.getItem('token')
  const response = await fetch(`/api/v1${endpoint}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Ошибка скачивания' }))
    throw new Error(error.error || 'Ошибка скачивания отчёта')
  }

  const disposition = response.headers.get('Content-Disposition')
  let filename = fallbackFilename
  if (disposition) {
    const match = disposition.match(/filename="?([^";\n]+)"?/)
    if (match) filename = match[1]
  }

  const blob = await response.blob()
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}
