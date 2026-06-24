import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { X, Wrench } from 'lucide-react'

const maintenanceSchema = z.object({
  description: z.string().min(10, 'Описание минимум 10 символов'),
  next_maintenance: z.string().min(1, 'Укажите дату следующего ТО'),
})

type MaintenanceData = z.infer<typeof maintenanceSchema>

interface Props {
  equipmentId: string
  equipmentName: string
  onClose: () => void
  onSuccess: () => void
}

export function MaintenanceModal({ equipmentId, equipmentName, onClose, onSuccess }: Props) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<MaintenanceData>({
    resolver: zodResolver(maintenanceSchema),
  })

  const onSubmit = async (data: MaintenanceData) => {
    setIsSubmitting(true)
    try {
      await api.post(`/equipment/${equipmentId}/maintenance`, data)
      toast.success(`ТО для "${equipmentName}" зафиксировано! ✅`)
      onSuccess()
      onClose()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка сохранения ТО')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div className="flex items-center justify-between p-6 border-b">
          <div className="flex items-center gap-3">
            <Wrench className="w-6 h-6 text-sky-500" />
            <h2 className="text-xl font-bold">Фиксация ТО</h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6">
          <div className="mb-4 p-3 bg-sky-50 rounded-lg">
            <p className="text-sm text-gray-600">Оборудование:</p>
            <p className="font-semibold text-sky-700">{equipmentName}</p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                Описание выполненных работ
              </label>
              <textarea
                {...register('description')}
                rows={4}
                placeholder="Замена подшипников, смазка направляющих, проверка тормозной системы..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
              />
              {errors.description && (
                <p className="text-sm text-red-500 mt-1">{errors.description.message}</p>
              )}
            </div>

            <Input
              label="Дата следующего ТО"
              type="date"
              error={errors.next_maintenance?.message}
              {...register('next_maintenance')}
            />

            <div className="flex gap-3 pt-4">
              <Button type="button" variant="secondary" onClick={onClose} className="flex-1">
                Отмена
              </Button>
              <Button type="submit" disabled={isSubmitting} className="flex-1">
                {isSubmitting ? 'Сохранение...' : '✅ Зафиксировать ТО'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}