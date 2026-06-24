import { useEffect, useState } from 'react'
import { useForm, SubmitHandler } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Droplets, AlertTriangle, CheckCircle, FlaskConical } from 'lucide-react'
import { clsx } from 'clsx'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'

// Схема валидации (нормы СанПиН)
const waterSchema = z.object({
  zone_id: z.string().min(1, 'Выберите зону'),
  ph: z.number().min(0, 'pH не может быть отрицательным').max(14, 'pH не может быть больше 14'),
  chlorine: z.number().min(0, 'Хлор не может быть отрицательным'),
  turbidity: z.number().min(0, 'Мутность не может быть отрицательной'),
})

type WaterData = z.infer<typeof waterSchema>

interface Zone {
  id: string
  name: string
}

interface Measurement {
  id: string
  zone_id: string
  ph: number
  chlorine: number
  turbidity: number
  is_normal: boolean
  violations: string[]
  measured_at: string
  zone: { name: string }
  technician: { full_name: string }
}

// Нормы СанПиН для отображения
const SANPIN = {
  phMin: 7.2,
  phMax: 7.6,
  chlorineMax: 0.5,
  turbidityMax: 1.5,
}

export function WaterQuality() {
  const [zones, setZones] = useState<Zone[]>([])
  const [measurements, setMeasurements] = useState<Measurement[]>([])
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [loading, setLoading] = useState(true)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<WaterData>({
    resolver: zodResolver(waterSchema),
  })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [zonesRes, measurementsRes] = await Promise.allSettled([
        api.get('/zones'),
        api.get('/water/measurements?limit=50'),
      ])

      if (zonesRes.status === 'fulfilled') {
        setZones(zonesRes.value.data.zones || [])
      }

      if (measurementsRes.status === 'fulfilled') {
        setMeasurements(measurementsRes.value.data.measurements || [])
      }
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  // Явная типизация onSubmit
  const onSubmit: SubmitHandler<WaterData> = async (data) => {
    setIsSubmitting(true)
    try {
      const res = await api.post('/water/measurements', data)
      
      if (res.data.is_normal) {
        toast.success('Замер сохранён. Показатели в норме! ✅')
      } else {
        toast.error('⚠️ Нарушение норм СанПиН!', {
          description: res.data.violations?.join(', ') || 'Проверьте показатели',
        })
      }

      // Добавляем новый замер в начало списка
      if (res.data.measurement) {
        setMeasurements([res.data.measurement, ...measurements])
      }
      
      reset()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка сохранения замера')
    } finally {
      setIsSubmitting(false)
    }
  }

  // Проверка показателей для цветовой индикации
  const getPhStatus = (ph: number) => {
    if (ph < SANPIN.phMin || ph > SANPIN.phMax) return 'text-red-600 bg-red-50'
    return 'text-green-600 bg-green-50'
  }

  const getChlorineStatus = (val: number) => {
    if (val > SANPIN.chlorineMax) return 'text-red-600 bg-red-50'
    return 'text-green-600 bg-green-50'
  }

  const getTurbidityStatus = (val: number) => {
    if (val > SANPIN.turbidityMax) return 'text-red-600 bg-red-50'
    return 'text-green-600 bg-green-50'
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Загрузка данных...</p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Droplets className="w-8 h-8 text-blue-500" />
          Журнал водоподготовки
        </h1>
        <p className="text-gray-500 mt-1">
          Контроль качества воды по нормам СанПиН
        </p>
      </div>

      {/* Нормы СанПиН (справка) */}
      <Card className="mb-6 bg-blue-50 border-blue-200">
        <CardContent className="py-4">
          <div className="flex items-center gap-2 text-blue-800 font-semibold mb-2">
            <FlaskConical className="w-5 h-5" />
            Нормы СанПиН для бассейнов:
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div>
              <span className="text-gray-600">pH:</span>{' '}
              <span className="font-bold">{SANPIN.phMin} – {SANPIN.phMax}</span>
            </div>
            <div>
              <span className="text-gray-600">Свободный хлор:</span>{' '}
              <span className="font-bold">≤ {SANPIN.chlorineMax} мг/л</span>
            </div>
            <div>
              <span className="text-gray-600">Мутность:</span>{' '}
              <span className="font-bold">≤ {SANPIN.turbidityMax} ЕМФ</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Форма ввода замера */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Новый замер</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {/* Выбор зоны */}
              <div>
                <label className="block text-sm font-medium mb-2">Зона</label>
                <select
                  {...register('zone_id')}
                  className={clsx(
                    'w-full px-3 py-2 border rounded-md text-sm',
                    'focus:outline-none focus:ring-2 focus:ring-blue-500',
                    errors.zone_id ? 'border-red-500' : 'border-gray-300'
                  )}
                >
                  <option value="">-- Выберите зону --</option>
                  {zones.map(zone => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
                {errors.zone_id && (
                  <p className="text-sm text-red-500 mt-1">{errors.zone_id.message}</p>
                )}
              </div>

              <Input
                label="pH уровень"
                type="number"
                step="0.1"
                placeholder="7.4"
                error={errors.ph?.message}
                {...register('ph', { valueAsNumber: true })}
              />

              <Input
                label="Хлор (мг/л)"
                type="number"
                step="0.01"
                placeholder="0.3"
                error={errors.chlorine?.message}
                {...register('chlorine', { valueAsNumber: true })}
              />

              <Input
                label="Мутность (ЕМФ)"
                type="number"
                step="0.1"
                placeholder="0.8"
                error={errors.turbidity?.message}
                {...register('turbidity', { valueAsNumber: true })}
              />

              <Button type="submit" size="lg" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? 'Сохранение...' : '💧 Сохранить замер'}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Таблица замеров */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>История замеров</CardTitle>
          </CardHeader>
          <CardContent>
            {measurements.length === 0 ? (
              <p className="text-center py-8 text-gray-500">
                Замеров пока нет
              </p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-gray-500">
                      <th className="pb-3">Дата</th>
                      <th className="pb-3">Зона</th>
                      <th className="pb-3">pH</th>
                      <th className="pb-3">Хлор</th>
                      <th className="pb-3">Мутность</th>
                      <th className="pb-3">Статус</th>
                    </tr>
                  </thead>
                  <tbody>
                    {measurements.map(m => (
                      <tr key={m.id} className="border-b hover:bg-gray-50">
                        <td className="py-3 text-sm">
                          {format(new Date(m.measured_at), 'dd MMM, HH:mm', {
                            locale: ru,
                          })}
                        </td>
                        <td className="py-3 text-sm font-medium">
                          {m.zone?.name || '—'}
                        </td>
                        <td className="py-3">
                          <span className={clsx('px-2 py-1 rounded text-xs font-bold', getPhStatus(m.ph))}>
                            {m.ph.toFixed(2)}
                          </span>
                        </td>
                        <td className="py-3">
                          <span className={clsx('px-2 py-1 rounded text-xs font-bold', getChlorineStatus(m.chlorine))}>
                            {m.chlorine.toFixed(2)}
                          </span>
                        </td>
                        <td className="py-3">
                          <span className={clsx('px-2 py-1 rounded text-xs font-bold', getTurbidityStatus(m.turbidity))}>
                            {m.turbidity.toFixed(2)}
                          </span>
                        </td>
                        <td className="py-3">
                          {m.is_normal ? (
                            <span className="flex items-center gap-1 text-xs text-green-700">
                              <CheckCircle className="w-4 h-4" /> Норма
                            </span>
                          ) : (
                            <span className="flex items-center gap-1 text-xs text-red-700">
                              <AlertTriangle className="w-4 h-4" /> Нарушение
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}