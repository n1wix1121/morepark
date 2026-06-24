import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Ticket, MapPin, Users, CheckCircle } from 'lucide-react'
import { clsx } from 'clsx'

const sellSchema = z.object({
  ticket_type_id: z.string().min(1, 'Выберите тип билета'),
  zone_id: z.string().min(1, 'Выберите зону'),
  customer_name: z.string().min(2, 'Введите имя клиента'),
  customer_phone: z.string().min(10, 'Введите телефон'),
  payment_method: z.enum(['cash', 'card']),
})

type SellData = z.infer<typeof sellSchema>

interface Zone {
  id: string
  name: string
  capacity: number
  current_count: number
  description?: string
}

interface TicketType {
  id: string
  type: string
  name: string
  price: number
  duration_hours: number
}

function getLoadPercent(zone: Zone): number {
  if (!zone.capacity || zone.capacity === 0) return 0
  return (zone.current_count / zone.capacity) * 100
}

function isZoneAvailable(zone: Zone, guests: number): boolean {
  return zone.capacity - zone.current_count >= guests
}

export function SellTicket() {
  const [zones, setZones] = useState<Zone[]>([])
  const [ticketTypes, setTicketTypes] = useState<TicketType[]>([])
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [lastSale, setLastSale] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<SellData>({
    resolver: zodResolver(sellSchema),
  })

  const selectedZoneId = watch('zone_id')
  const selectedTicketTypeId = watch('ticket_type_id')
  const selectedZone = zones.find(z => z.id === selectedZoneId)
  const selectedTicketType = ticketTypes.find(t => t.id === selectedTicketTypeId)
  const guestCount = selectedTicketType?.type === 'group' ? 5 : 1

  useEffect(() => {
    loadData()
    const interval = setInterval(loadZones, 3000)
    return () => clearInterval(interval)
  }, [])

  const loadZones = async () => {
    try {
      const res = await api.get('/zones')
      setZones(res.data.zones || [])
    } catch (err) {
      console.error('Failed to load zones:', err)
    }
  }

  const loadData = async () => {
    setLoading(true)
    try {
      const [zonesRes, typesRes] = await Promise.allSettled([
        api.get('/zones'),
        api.get('/ticket-types'),
      ])

      if (zonesRes.status === 'fulfilled') {
        setZones(zonesRes.value.data.zones || [])
      } else {
        console.error('Failed to load zones:', zonesRes.reason)
      }

      if (typesRes.status === 'fulfilled') {
        setTicketTypes(typesRes.value.data.ticket_types || [])
      } else {
        console.error('Failed to load ticket types:', typesRes.reason)
      }
    } catch (err) {
      console.error('Unexpected error:', err)
    } finally {
      setLoading(false)
    }
  }

  const onSubmit = async (data: SellData) => {
    setIsSubmitting(true)
    try {
      const res = await api.post('/tickets/sell', data)
      setLastSale(res.data.sale)
      
      toast.success('Билет успешно продан!', {
        description: selectedZone 
          ? `${selectedZone.name}: +${guestCount} чел. → ${selectedZone.current_count + guestCount}/${selectedZone.capacity}`
          : 'Билет продан',
      })
      
      reset()
      loadZones()
    } catch (err: any) {
      const error = err.response?.data
      if (error?.code === 'ZONE_FULL') {
        toast.error('Зона заполнена!', {
          description: 'Продажа невозможна — достигнут лимит вместимости',
        })
      } else {
        toast.error(error?.error || 'Ошибка продажи')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const getZoneColor = (zone: Zone) => {
    const percent = getLoadPercent(zone)
    if (percent >= 100) return 'border-red-500 bg-red-50'
    if (percent >= 80) return 'border-yellow-500 bg-yellow-50'
    return 'border-green-500 bg-green-50'
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
          <Ticket className="w-8 h-8 text-sky-500" />
          Продажа билетов
        </h1>
        <p className="text-gray-500 mt-1">
          Выберите зону, тип билета и оформите продажу
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Зоны */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MapPin className="w-5 h-5" />
              Зоны аквапарка
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {zones.length === 0 ? (
              <p className="text-center py-8 text-gray-500">Нет доступных зон</p>
            ) : (
              zones.map(zone => {
                const loadPercent = getLoadPercent(zone)
                const available = isZoneAvailable(zone, guestCount)
                const isSelected = selectedZoneId === zone.id
                
                return (
                  <div
                    key={zone.id}
                    onClick={() => {
                      if (available) setValue('zone_id', zone.id)
                    }}
                    className={clsx(
                      'block p-4 border-2 rounded-lg cursor-pointer transition-all',
                      getZoneColor(zone),
                      isSelected && 'ring-2 ring-sky-500 border-sky-500',
                      !available && 'opacity-50 cursor-not-allowed'
                    )}
                  >
                    <input
                      type="radio"
                      value={zone.id}
                      checked={isSelected}
                      onChange={() => setValue('zone_id', zone.id)}
                      disabled={!available}
                      className="sr-only"
                    />
                    <div className="flex justify-between items-start mb-2">
                      <div className="font-semibold">{zone.name}</div>
                      {!available && (
                        <span className="text-xs bg-red-500 text-white px-2 py-1 rounded">
                          ЗАПОЛНЕНА
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <Users className="w-4 h-4" />
                      <span className="font-bold">
                        {zone.current_count} / {zone.capacity}
                      </span>
                      <span className="text-gray-500">
                        ({loadPercent.toFixed(0)}%)
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-1.5 mt-2">
                      <div
                        className={clsx(
                          'h-1.5 rounded-full transition-all',
                          loadPercent >= 100 ? 'bg-red-500' :
                          loadPercent >= 80 ? 'bg-yellow-500' : 'bg-green-500'
                        )}
                        style={{ width: `${Math.min(loadPercent, 100)}%` }}
                      />
                    </div>
                  </div>
                )
              })
            )}
          </CardContent>
        </Card>

        {/* Форма продажи */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Оформление продажи</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {/* Тип билета */}
              <div>
                <label className="block text-sm font-medium mb-2">Тип билета</label>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  {ticketTypes.map(type => {
                    const isSelected = selectedTicketTypeId === type.id
                    
                    return (
                      <div
                        key={type.id}
                        onClick={() => setValue('ticket_type_id', type.id)}
                        className={clsx(
                          'block p-4 border-2 rounded-lg cursor-pointer transition-all',
                          isSelected ? 'border-sky-500 bg-sky-50' : 'border-gray-200 hover:border-sky-300'
                        )}
                      >
                        <input
                          type="radio"
                          value={type.id}
                          checked={isSelected}
                          onChange={() => setValue('ticket_type_id', type.id)}
                          className="sr-only"
                        />
                        <div className="font-semibold">{type.name}</div>
                        <div className="text-2xl font-bold text-sky-500 mt-1">
                          {type.price.toLocaleString()} ₽
                        </div>
                        <div className="text-xs text-gray-500">
                          {type.duration_hours} ч · {type.type === 'group' ? '5 человек' : '1 человек'}
                        </div>
                      </div>
                    )
                  })}
                </div>
                {errors.ticket_type_id && (
                  <p className="text-sm text-red-500 mt-1">{errors.ticket_type_id.message}</p>
                )}
              </div>

              {/* Данные клиента */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="Имя клиента"
                  placeholder="Иванов Иван Иванович"
                  error={errors.customer_name?.message}
                  {...register('customer_name')}
                />
                <Input
                  label="Телефон"
                  placeholder="+7 (900) 123-45-67"
                  error={errors.customer_phone?.message}
                  {...register('customer_phone')}
                />
              </div>

              {/* Способ оплаты */}
              <div>
                <label className="block text-sm font-medium mb-2">Способ оплаты</label>
                <div className="flex gap-3">
                  <label className="flex-1">
                    <input
                      type="radio"
                      value="cash"
                      {...register('payment_method')}
                      className="sr-only peer"
                    />
                    <div className="p-4 border-2 rounded-lg text-center cursor-pointer peer-checked:border-sky-500 peer-checked:bg-sky-50">
                      💵 Наличные
                    </div>
                  </label>
                  <label className="flex-1">
                    <input
                      type="radio"
                      value="card"
                      {...register('payment_method')}
                      className="sr-only peer"
                    />
                    <div className="p-4 border-2 rounded-lg text-center cursor-pointer peer-checked:border-sky-500 peer-checked:bg-sky-50">
                      💳 Карта
                    </div>
                  </label>
                </div>
                {errors.payment_method && (
                  <p className="text-sm text-red-500 mt-1">{errors.payment_method.message}</p>
                )}
              </div>

              {/* Итог */}
              {selectedZone && (
                <div className="p-4 bg-sky-50 border border-sky-200 rounded-lg">
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-600">Зона:</span>
                    <span className="font-semibold">{selectedZone.name}</span>
                  </div>
                  <div className="flex justify-between items-center mt-2">
                    <span className="text-sm text-gray-600">Свободных мест:</span>
                    <span className={clsx(
                      'font-bold',
                      selectedZone.capacity - selectedZone.current_count < guestCount ? 'text-red-600' : 'text-green-600'
                    )}>
                      {selectedZone.capacity - selectedZone.current_count}
                      {guestCount > 1 && ` (нужно ${guestCount})`}
                    </span>
                  </div>
                </div>
              )}

              {/* Кнопка продажи */}
              <Button
                type="submit"
                size="lg"
                className="w-full"
                disabled={
                  isSubmitting ||
                  !selectedZoneId ||
                  !selectedTicketTypeId ||
                  !watch('customer_name') ||
                  !watch('customer_phone') ||
                  !watch('payment_method')
                }
              >
                {isSubmitting ? 'Продажа...' : '💰 Продать билет'}
              </Button>
            </form>

            {/* Последняя продажа */}
            {lastSale && (
              <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
                <div className="flex items-center gap-2 text-green-700 font-semibold mb-2">
                  <CheckCircle className="w-5 h-5" />
                  Последняя продажа
                </div>
                <div className="text-sm space-y-1">
                  <div>
                    <strong>Сумма:</strong> {lastSale.amount?.toLocaleString()} ₽
                  </div>
                  <div>
                    <strong>Оплата:</strong>{' '}
                    {lastSale.payment_method === 'cash' ? 'Наличные' : 'Карта'}
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}