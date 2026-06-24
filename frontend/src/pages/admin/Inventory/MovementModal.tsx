import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { X, Package } from 'lucide-react'

const movementSchema = z.object({
  type: z.enum(['in', 'out']),
  quantity: z.coerce.number().min(0.1, 'Количество должно быть больше 0'),
  reason: z.string().min(5, 'Укажите причину (минимум 5 символов)'),
})

type MovementData = z.infer<typeof movementSchema>

interface Props {
  inventoryId: string
  itemName: string
  currentQuantity: number
  type: 'in' | 'out'
  onClose: () => void
  onSuccess: () => void
}

export function MovementModal({ inventoryId, itemName, currentQuantity, type, onClose, onSuccess }: Props) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<MovementData>({
    resolver: zodResolver(movementSchema),
    defaultValues: { type },
  })

  const onSubmit = async (data: MovementData) => {
    if (type === 'out' && data.quantity > currentQuantity) {
      toast.error(`Недостаточно товара! Доступно: ${currentQuantity}`)
      return
    }

    setIsSubmitting(true)
    try {
      await api.post(`/inventory/${inventoryId}/move`, data)
      toast.success(
        type === 'in' 
          ? `Приход "${itemName}" зафиксирован ✅` 
          : `Расход "${itemName}" зафиксирован ✅`
      )
      onSuccess()
      onClose()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка операции')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div className="flex items-center justify-between p-6 border-b">
          <div className="flex items-center gap-3">
            <Package className="w-6 h-6 text-sky-500" />
            <h2 className="text-xl font-bold">
              {type === 'in' ? 'Приход товара' : 'Расход товара'}
            </h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6">
          <div className="mb-4 p-3 bg-sky-50 rounded-lg">
            <p className="text-sm text-gray-600">Товар:</p>
            <p className="font-semibold text-sky-700">{itemName}</p>
            <p className="text-xs text-gray-500 mt-1">
              Текущий остаток: <strong>{currentQuantity}</strong>
            </p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <Input
              label="Количество"
              type="number"
              step="0.1"
              placeholder="0.0"
              error={errors.quantity?.message}
              {...register('quantity')}
            />

            <div>
              <label className="block text-sm font-medium mb-2">
                Причина / Описание
              </label>
              <textarea
                {...register('reason')}
                rows={3}
                placeholder={
                  type === 'in' 
                    ? 'Закупка у поставщика ХимПром' 
                    : 'Использовано для обработки бассейна'
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
              />
              {errors.reason && (
                <p className="text-sm text-red-500 mt-1">{errors.reason.message}</p>
              )}
            </div>

            <div className="flex gap-3 pt-4">
              <Button type="button" variant="secondary" onClick={onClose} className="flex-1">
                Отмена
              </Button>
              <Button type="submit" disabled={isSubmitting} className="flex-1">
                {isSubmitting ? 'Сохранение...' : type === 'in' ? '✅ Оприходовать' : '📤 Списать'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}