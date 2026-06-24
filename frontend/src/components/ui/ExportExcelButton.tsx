import { useState } from 'react'
import { FileSpreadsheet } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/Button'
import { downloadExcelReport } from '@/utils/downloadReport'

interface Props {
  endpoint: string
  filename: string
  label?: string
  size?: 'sm' | 'md' | 'lg'
}

export function ExportExcelButton({
  endpoint,
  filename,
  label = 'Скачать Excel',
  size = 'md',
}: Props) {
  const [loading, setLoading] = useState(false)

  const handleExport = async () => {
    setLoading(true)
    try {
      await downloadExcelReport(endpoint, filename)
      toast.success('Отчёт скачан')
    } catch (err: any) {
      toast.error(err.message || 'Ошибка скачивания')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Button
      variant="secondary"
      size={size}
      onClick={handleExport}
      disabled={loading}
      className="border-green-200 bg-green-50 text-green-700 hover:bg-green-100"
    >
      <FileSpreadsheet className="w-4 h-4 mr-2" />
      {loading ? 'Формирование...' : label}
    </Button>
  )
}
