import { useEffect, useState } from 'react'
import { useForm, SubmitHandler } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { 
  Ticket, 
  MapPin, 
  Users, 
  Calendar, 
  User, 
  Phone, 
  Mail, 
  CheckCircle,
  QrCode,
  ArrowLeft,
  Waves
} from 'lucide-react'
import { clsx } from 'clsx'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'

// Схема валидации покупки
const purchaseSchema = z.object({
  ticket_type_id: z.string().min(1, 'Выберите тип билета'),
  zone_id: z.string().min(1, 'Выберите зону'),
  datetime: z.string().min(1, 'Выберите дату и время'),
  quantity: z.coerce.number().min(1, 'Минимум 1 билет').max(10, 'Максимум 10 билетов'),
  customer_name: z.string().min(2, 'Введите имя'),
  customer_phone: z.string().min(10, 'Введите телефон'),
  customer_email: z.string().email('Неверный email'),
})

type PurchaseData = z.infer<typeof purchaseSchema>

interface Zone {
  id: string
  name: string
  capacity: number
  current_count: number
  description: string
  is_available: boolean
  load_percent: number
}

interface TicketType {
  id: string
  type: string
  name: string
  price: number
  duration_hours: number
  description: string
}

interface PurchaseResponse {
  ticket: {
    id: string
    ticket_number: string
    customer_name: string
    zone: { name: string }
    valid_from: string
    valid_until: string
    quantity: number
  }
  ticket_number: string
  qr_code_base64: string
  amount: number
  message: string
}

// Шаги покупки
type Step = 'select' | 'details' | 'success'

export function BuyTicket() {
  const [step, setStep] = useState<Step>('select')
  const [zones, setZones] = useState<Zone[]>([])
  const [ticketTypes, setTicketTypes] = useState<TicketType[]>([])
  const [loading, setLoading] = useState(true)
  const [purchaseResult, setPurchaseResult] = useState<PurchaseResponse | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<PurchaseData>({
    resolver: zodResolver(purchaseSchema),
    defaultValues: {
      quantity: 1,
    },
  })

  const selectedZoneId = watch('zone_id')
  const selectedTicketTypeId = watch('ticket_type_id')
  const selectedZone = zones.find(z => z.id === selectedZoneId)
  const selectedTicketType = ticketTypes.find(t => t.id === selectedTicketTypeId)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [zonesRes, typesRes] = await Promise.allSettled([
        api.get('/public/zones'),
        api.get('/public/ticket-types'),
      ])

      if (zonesRes.status === 'fulfilled') {
        setZones(zonesRes.value.data.zones || [])
      }

      if (typesRes.status === 'fulfilled') {
        setTicketTypes(typesRes.value.data.ticket_types || [])
      }
    } catch (err) {
      console.error(err)
      toast.error('Ошибка загрузки данных')
    } finally {
      setLoading(false)
    }
  }

  const onSubmit: SubmitHandler<PurchaseData> = async (data) => {
    setIsSubmitting(true)
    try {
      const datetime = new Date(data.datetime).toISOString()

      const availRes = await api.post('/public/check-availability', {
        zone_id: data.zone_id,
        datetime,
        quantity: data.quantity,
      })

      if (!availRes.data.available) {
        toast.error(availRes.data.message || 'Недостаточно мест на выбранное время')
        return
      }

      const res = await api.post('/public/tickets', { ...data, datetime })
      setPurchaseResult(res.data)
      setStep('success')
      toast.success('Билет успешно забронирован! 🎉')
    } catch (err: any) {
      const error = err.response?.data
      if (error?.code === 'NOT_ENOUGH_SLOTS') {
        toast.error('Недостаточно мест на выбранное время')
      } else {
        toast.error(error?.error || 'Ошибка бронирования')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const getLoadColor = (percent: number) => {
    if (percent >= 100) return 'bg-red-500'
    if (percent >= 80) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  const resetPurchase = () => {
    setStep('select')
    setPurchaseResult(null)
    setValue('zone_id', '')
    setValue('ticket_type_id', '')
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-sky-400 to-blue-600 flex items-center justify-center">
        <div className="bg-white rounded-lg p-8 shadow-xl">
          <p className="text-gray-500">Загрузка...</p>
        </div>
      </div>
    )
  }

  // === ШАГ 1: Выбор зоны и типа билета ===
  if (step === 'select') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-sky-400 to-blue-600 py-8 px-4">
        <div className="max-w-6xl mx-auto">
          {/* Шапка */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center gap-3 bg-white/20 backdrop-blur-sm px-6 py-3 rounded-full mb-4">
              <Waves className="w-6 h-6 text-white" />
              <h1 className="text-2xl font-bold text-white">Море Парк</h1>
            </div>
            <h2 className="text-4xl font-bold text-white mb-2">
              Покупка билетов онлайн
            </h2>
            <p className="text-white/80">
              Выберите зону и тип билета для бронирования
            </p>
          </div>

          {/* Зоны */}
          <div className="mb-8">
            <h3 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
              <MapPin className="w-5 h-5" />
              Выберите зону
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {zones.map(zone => {
                const isSelected = selectedZoneId === zone.id
                const available = zone.current_count < zone.capacity

                return (
                  <div
                    key={zone.id}
                    onClick={() => available && setValue('zone_id', zone.id)}
                    className={clsx(
                      'bg-white rounded-lg p-5 cursor-pointer transition-all',
                      'hover:shadow-xl hover:-translate-y-1',
                      isSelected ? 'ring-4 ring-white shadow-xl' : '',
                      !available && 'opacity-50 cursor-not-allowed'
                    )}
                  >
                    <div className="flex justify-between items-start mb-3">
                      <h4 className="font-bold text-lg">{zone.name}</h4>
                      {!available && (
                        <span className="text-xs bg-red-500 text-white px-2 py-1 rounded">
                          ЗАПОЛНЕНА
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-600 mb-3 line-clamp-2">
                      {zone.description}
                    </p>
                    <div className="flex items-center gap-2 text-sm mb-2">
                      <Users className="w-4 h-4 text-gray-400" />
                      <span className="font-semibold">
                        {zone.current_count} / {zone.capacity}
                      </span>
                      <span className="text-gray-500">
                        ({zone.load_percent.toFixed(0)}%)
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className={clsx('h-2 rounded-full transition-all', getLoadColor(zone.load_percent))}
                        style={{ width: `${Math.min(zone.load_percent, 100)}%` }}
                      />
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Типы билетов */}
          <div className="mb-8">
            <h3 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
              <Ticket className="w-5 h-5" />
              Выберите тип билета
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {ticketTypes.map(type => {
                const isSelected = selectedTicketTypeId === type.id

                return (
                  <div
                    key={type.id}
                    onClick={() => setValue('ticket_type_id', type.id)}
                    className={clsx(
                      'bg-white rounded-lg p-5 cursor-pointer transition-all',
                      'hover:shadow-xl hover:-translate-y-1',
                      isSelected ? 'ring-4 ring-white shadow-xl' : ''
                    )}
                  >
                    <h4 className="font-bold text-lg mb-2">{type.name}</h4>
                    <p className="text-sm text-gray-600 mb-3">
                      {type.description}
                    </p>
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-3xl font-bold text-sky-500">
                          {type.price.toLocaleString()} ₽
                        </div>
                        <div className="text-xs text-gray-500">
                          <Calendar className="w-3 h-3 inline mr-1" />
                          {type.duration_hours} часов
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Кнопка далее */}
          <div className="text-center">
            <Button
              size="lg"
              onClick={() => {
                if (!selectedZoneId || !selectedTicketTypeId) {
                  toast.error('Выберите зону и тип билета')
                  return
                }
                setStep('details')
              }}
              className="px-12"
              disabled={!selectedZoneId || !selectedTicketTypeId}
            >
              Далее →
            </Button>
          </div>
        </div>
      </div>
    )
  }

  // === ШАГ 2: Данные покупателя ===
  if (step === 'details') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-sky-400 to-blue-600 py-8 px-4">
        <div className="max-w-2xl mx-auto">
          <button
            onClick={() => setStep('select')}
            className="flex items-center gap-2 text-white mb-6 hover:underline"
          >
            <ArrowLeft className="w-4 h-4" />
            Назад к выбору
          </button>

          <Card>
            <CardHeader>
              <CardTitle className="text-2xl">Данные для бронирования</CardTitle>
            </CardHeader>
            <CardContent>
              {/* Сводка выбора */}
              <div className="bg-sky-50 border border-sky-200 rounded-lg p-4 mb-6">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <p className="text-gray-600">Зона:</p>
                    <p className="font-bold text-sky-700">{selectedZone?.name}</p>
                  </div>
                  <div>
                    <p className="text-gray-600">Тип билета:</p>
                    <p className="font-bold text-sky-700">{selectedTicketType?.name}</p>
                  </div>
                  <div>
                    <p className="text-gray-600">Цена:</p>
                    <p className="font-bold text-sky-700">
                      {selectedTicketType?.price.toLocaleString()} ₽
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-600">Длительность:</p>
                    <p className="font-bold text-sky-700">
                      {selectedTicketType?.duration_hours} часов
                    </p>
                  </div>
                </div>
              </div>

              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                {/* Дата и время */}
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Дата и время посещения
                  </label>
                  <input
                    type="datetime-local"
                    {...register('datetime')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  />
                  {errors.datetime && (
                    <p className="text-sm text-red-500 mt-1">{errors.datetime.message}</p>
                  )}
                </div>

                {/* Количество */}
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Количество билетов
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={10}
                    {...register('quantity')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  />
                  {errors.quantity && (
                    <p className="text-sm text-red-500 mt-1">{errors.quantity.message}</p>
                  )}
                </div>

                {/* Имя */}
                <div>
                  <label className="block text-sm font-medium mb-2">
                    ФИО
                  </label>
                  <div className="relative">
                    <User className="w-4 h-4 absolute left-3 top-3 text-gray-400" />
                    <input
                      type="text"
                      placeholder="Иванов Иван Иванович"
                      {...register('customer_name')}
                      className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                    />
                  </div>
                  {errors.customer_name && (
                    <p className="text-sm text-red-500 mt-1">{errors.customer_name.message}</p>
                  )}
                </div>

                {/* Телефон */}
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Телефон
                  </label>
                  <div className="relative">
                    <Phone className="w-4 h-4 absolute left-3 top-3 text-gray-400" />
                    <input
                      type="tel"
                      placeholder="+7 (900) 123-45-67"
                      {...register('customer_phone')}
                      className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                    />
                  </div>
                  {errors.customer_phone && (
                    <p className="text-sm text-red-500 mt-1">{errors.customer_phone.message}</p>
                  )}
                </div>

                {/* Email */}
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Email
                  </label>
                  <div className="relative">
                    <Mail className="w-4 h-4 absolute left-3 top-3 text-gray-400" />
                    <input
                      type="email"
                      placeholder="ivan@example.com"
                      {...register('customer_email')}
                      className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                    />
                  </div>
                  {errors.customer_email && (
                    <p className="text-sm text-red-500 mt-1">{errors.customer_email.message}</p>
                  )}
                </div>

                {/* Итого */}
                <div className="bg-sky-50 border border-sky-200 rounded-lg p-4">
                  <div className="flex justify-between items-center">
                    <span className="text-lg font-semibold">Итого к оплате:</span>
                    <span className="text-3xl font-bold text-sky-600">
                      {((selectedTicketType?.price || 0) * (watch('quantity') || 1)).toLocaleString()} ₽
                    </span>
                  </div>
                </div>

                <div className="flex gap-3">
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setStep('select')}
                    className="flex-1"
                  >
                    Назад
                  </Button>
                  <Button type="submit" size="lg" disabled={isSubmitting} className="flex-1">
                    {isSubmitting ? 'Бронирование...' : '🎫 Забронировать'}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  // === ШАГ 3: Успех + QR ===
  if (step === 'success' && purchaseResult) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-sky-400 to-blue-600 py-8 px-4">
        <div className="max-w-md mx-auto">
          <Card>
            <CardContent className="p-8 text-center">
              <div className="inline-flex items-center justify-center w-20 h-20 bg-green-100 rounded-full mb-4">
                <CheckCircle className="w-12 h-12 text-green-600" />
              </div>

              <h2 className="text-2xl font-bold mb-2">Билет забронирован!</h2>
              <p className="text-gray-600 mb-6">
                {purchaseResult.message}
              </p>

              {/* Номер билета */}
              <div className="bg-sky-50 border-2 border-dashed border-sky-300 rounded-lg p-4 mb-6">
                <p className="text-xs text-gray-600 mb-1">Номер билета:</p>
                <p className="text-2xl font-bold text-sky-700 font-mono">
                  {purchaseResult.ticket_number}
                </p>
              </div>

              {/* QR-код */}
              {purchaseResult.qr_code_base64 && (
                <div className="mb-6">
                  <p className="text-sm text-gray-600 mb-2">QR-код для входа:</p>
                  <div className="inline-block bg-white p-4 rounded-lg shadow-md">
                    <img
                      src={`data:image/png;base64,${purchaseResult.qr_code_base64}`}
                      alt="QR Code"
                      className="w-48 h-48"
                    />
                  </div>
                </div>
              )}

              {/* Детали */}
              <div className="bg-gray-50 rounded-lg p-4 mb-6 text-left space-y-2">
                <div className="flex justify-between">
                  <span className="text-gray-600">Зона:</span>
                  <span className="font-medium">{purchaseResult.ticket.zone.name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Посетитель:</span>
                  <span className="font-medium">{purchaseResult.ticket.customer_name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Время:</span>
                  <span className="font-medium">
                    {format(new Date(purchaseResult.ticket.valid_from), 'dd MMM, HH:mm', { locale: ru })}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">До:</span>
                  <span className="font-medium">
                    {format(new Date(purchaseResult.ticket.valid_until), 'dd MMM, HH:mm', { locale: ru })}
                  </span>
                </div>
                <div className="flex justify-between pt-2 border-t">
                  <span className="text-gray-600 font-semibold">Оплачено:</span>
                  <span className="font-bold text-green-600">
                    {purchaseResult.amount.toLocaleString()} ₽
                  </span>
                </div>
              </div>

              <div className="space-y-3">
                <Button onClick={resetPurchase} className="w-full">
                  Купить ещё билет
                </Button>
                <p className="text-xs text-gray-500">
                  Покажите QR-код на входе в аквапарк
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return null
}