import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Wrench, AlertCircle } from 'lucide-react'
import { clsx } from 'clsx'
import { format, differenceInDays } from 'date-fns'
import { ru } from 'date-fns/locale'
import { MaintenanceModal } from './MaintenanceModal'

interface Equipment {
  id: string
  name: string
  serial_number: string
  zone_id: string | null
  status: 'working' | 'maintenance' | 'broken'
  last_maintenance: string | null
  next_maintenance: string | null
  created_at: string
  zone?: { name: string }
}

interface UpcomingMaintenance {
  equipment_id: string
  equipment_name: string
  serial_number: string
  next_maintenance: string
  days_until: number
  urgency_level: 'low' | 'medium' | 'high' | 'critical'
}

export function EquipmentList() {
  const [equipment, setEquipment] = useState<Equipment[]>([])
  const [upcoming, setUpcoming] = useState<UpcomingMaintenance[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedEquipment, setSelectedEquipment] = useState<Equipment | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [equipRes, upcomingRes] = await Promise.allSettled([
        api.get('/equipment'),
        api.get('/maintenance/upcoming?days=30'),
      ])

      if (equipRes.status === 'fulfilled') {
        setEquipment(equipRes.value.data.equipment || [])
      } else {
        console.error(equipRes.reason)
        toast.error('Ошибка загрузки оборудования')
      }

      if (upcomingRes.status === 'fulfilled') {
        setUpcoming(upcomingRes.value.data.upcoming_maintenance || [])
      }
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'working': return 'bg-green-100 text-green-700'
      case 'maintenance': return 'bg-yellow-100 text-yellow-700'
      case 'broken': return 'bg-red-100 text-red-700'
      default: return 'bg-gray-100 text-gray-700'
    }
  }

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'working': return 'Работает'
      case 'maintenance': return 'На ТО'
      case 'broken': return 'Сломано'
      default: return status
    }
  }

  const getUrgencyColor = (level: string) => {
    switch (level) {
      case 'critical': return 'bg-red-500 text-white'
      case 'high': return 'bg-orange-500 text-white'
      case 'medium': return 'bg-yellow-500 text-white'
      case 'low': return 'bg-blue-500 text-white'
      default: return 'bg-gray-500 text-white'
    }
  }

  const getUrgencyLabel = (level: string) => {
    switch (level) {
      case 'critical': return 'СРОЧНО'
      case 'high': return 'Высокий'
      case 'medium': return 'Средний'
      case 'low': return 'Низкий'
      default: return level
    }
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
          <Wrench className="w-8 h-8 text-sky-500" />
          Оборудование и ТО
        </h1>
        <p className="text-gray-500 mt-1">
          Картотека оборудования и календарь технического обслуживания
        </p>
      </div>

      {/* Напоминания о предстоящем ТО */}
      {upcoming.length > 0 && (
        <Card className="mb-6 border-orange-200 bg-orange-50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-orange-700">
              <AlertCircle className="w-5 h-5" />
              Предстоящее ТО ({upcoming.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {upcoming.map(item => (
                <div
                  key={item.equipment_id}
                  className="flex items-center justify-between p-3 bg-white rounded-lg border"
                >
                  <div className="flex items-center gap-3">
                    <div className={clsx('px-3 py-1 rounded text-xs font-bold', getUrgencyColor(item.urgency_level))}>
                      {getUrgencyLabel(item.urgency_level)}
                    </div>
                    <div>
                      <p className="font-semibold">{item.equipment_name}</p>
                      <p className="text-xs text-gray-500">{item.serial_number}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-bold">
                      {item.days_until === 0 ? 'Сегодня' : `Через ${item.days_until} дн.`}
                    </p>
                    <p className="text-xs text-gray-500">
                      {format(new Date(item.next_maintenance), 'dd MMM yyyy', { locale: ru })}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Список оборудования */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {equipment.map(eq => {
          const nextMaintenance = eq.next_maintenance ? new Date(eq.next_maintenance) : null
          const daysUntil = nextMaintenance ? differenceInDays(nextMaintenance, new Date()) : null

          return (
            <Card key={eq.id} className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <CardTitle className="text-lg">{eq.name}</CardTitle>
                    <p className="text-xs text-gray-500 mt-1">{eq.serial_number}</p>
                  </div>
                  <span className={clsx('px-2 py-1 rounded text-xs font-semibold', getStatusColor(eq.status))}>
                    {getStatusLabel(eq.status)}
                  </span>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {eq.zone && (
                  <div className="text-sm">
                    <span className="text-gray-500">Зона:</span>{' '}
                    <span className="font-medium">{eq.zone.name}</span>
                  </div>
                )}

                {eq.last_maintenance && (
                  <div className="text-sm">
                    <span className="text-gray-500">Последнее ТО:</span>{' '}
                    <span className="font-medium">
                      {format(new Date(eq.last_maintenance), 'dd MMM yyyy', { locale: ru })}
                    </span>
                  </div>
                )}

                {nextMaintenance && daysUntil !== null && (
                  <div className="text-sm">
                    <span className="text-gray-500">Следующее ТО:</span>{' '}
                    <div className="flex items-center gap-2 mt-1">
                      <span className="font-medium">
                        {format(nextMaintenance, 'dd MMM yyyy', { locale: ru })}
                      </span>
                      {daysUntil <= 7 && (
                        <span className={clsx(
                          'px-2 py-0.5 rounded text-xs font-bold',
                          daysUntil <= 0 ? 'bg-red-500 text-white' :
                          daysUntil <= 3 ? 'bg-orange-500 text-white' :
                          'bg-yellow-500 text-white'
                        )}>
                          {daysUntil <= 0 ? 'Сегодня' : `Через ${daysUntil} дн.`}
                        </span>
                      )}
                    </div>
                  </div>
                )}

                <Button
                  onClick={() => setSelectedEquipment(eq)}
                  className="w-full"
                  disabled={eq.status === 'broken'}
                >
                  <Wrench className="w-4 h-4 mr-2" />
                  Зафиксировать ТО
                </Button>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {equipment.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <Wrench className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">Оборудование не добавлено</p>
            <p className="text-sm text-gray-400 mt-2">Перезапустите backend — тестовые данные подгрузятся автоматически</p>
          </CardContent>
        </Card>
      )}

      {/* Модалка фиксации ТО */}
      {selectedEquipment && (
        <MaintenanceModal
          equipmentId={selectedEquipment.id}
          equipmentName={selectedEquipment.name}
          onClose={() => setSelectedEquipment(null)}
          onSuccess={loadData}
        />
      )}
    </div>
  )
}