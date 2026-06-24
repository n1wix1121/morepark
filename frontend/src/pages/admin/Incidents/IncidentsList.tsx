import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { 
  AlertTriangle, 
  Plus, 
  X, 
  CheckCircle, 
  Clock, 
  AlertCircle,
  Filter
} from 'lucide-react'
import { clsx } from 'clsx'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'

// Схема создания инцидента
const createSchema = z.object({
  zone_id: z.string().min(1, 'Выберите зону'),
  description: z.string().min(10, 'Описание минимум 10 символов'),
  severity: z.enum(['low', 'medium', 'high']),
})

type CreateData = z.infer<typeof createSchema>

// Схема изменения статуса
const statusSchema = z.object({
  status: z.enum(['open', 'in_progress', 'closed']),
  resolution: z.string().min(10, 'Описание решения минимум 10 символов').optional(),
})

type StatusData = z.infer<typeof statusSchema>

interface Incident {
  id: string
  zone_id: string
  lifeguard_id: string
  description: string
  severity: 'low' | 'medium' | 'high'
  status: 'open' | 'in_progress' | 'closed'
  resolved_by?: string
  resolved_at?: string
  resolution?: string
  created_at: string
  zone?: { name: string }
  lifeguard?: { full_name: string }
}

interface Zone {
  id: string
  name: string
}

const severityConfig = {
  low: { 
    label: 'Низкая', 
    color: 'bg-blue-100 text-blue-700',
    icon: AlertCircle 
  },
  medium: { 
    label: 'Средняя', 
    color: 'bg-yellow-100 text-yellow-700',
    icon: AlertTriangle 
  },
  high: { 
    label: 'Высокая', 
    color: 'bg-red-100 text-red-700',
    icon: AlertTriangle 
  },
}

const statusConfig = {
  open: { 
    label: 'Открыт', 
    color: 'bg-blue-100 text-blue-700',
    icon: AlertCircle 
  },
  in_progress: { 
    label: 'В работе', 
    color: 'bg-yellow-100 text-yellow-700',
    icon: Clock 
  },
  closed: { 
    label: 'Закрыт', 
    color: 'bg-green-100 text-green-700',
    icon: CheckCircle 
  },
}

export function IncidentsList() {
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [zones, setZones] = useState<Zone[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null)
  const [filterStatus, setFilterStatus] = useState<string>('active')
  const [filterSeverity, setFilterSeverity] = useState<string>('all')

  const createForm = useForm<CreateData>({
    resolver: zodResolver(createSchema),
  })

  const statusForm = useForm<StatusData>({
    resolver: zodResolver(statusSchema),
  })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [incidentsRes, zonesRes] = await Promise.allSettled([
        api.get('/incidents?limit=100'),
        api.get('/zones'),
      ])

      if (incidentsRes.status === 'fulfilled') {
        setIncidents(incidentsRes.value.data.incidents || [])
      } else {
        console.error(incidentsRes.reason)
        toast.error('Ошибка загрузки инцидентов')
      }

      if (zonesRes.status === 'fulfilled') {
        setZones(zonesRes.value.data.zones || [])
      }
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const onCreateSubmit = async (data: CreateData) => {
    try {
      const res = await api.post('/incidents', data)
      toast.success('Инцидент зарегистрирован! 🚨')
      setShowCreateModal(false)
      createForm.reset()
      if (res.data.incident) {
        setIncidents(prev => [res.data.incident, ...prev])
      }
      loadData()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка создания')
    }
  }

  const onStatusSubmit = async (data: StatusData) => {
    if (!selectedIncident) return

    // При закрытии обязательно описание решения
    if (data.status === 'closed' && (!data.resolution || data.resolution.length < 10)) {
      toast.error('При закрытии инцидента необходимо указать описание решения')
      return
    }

    try {
      await api.patch(`/incidents/${selectedIncident.id}/status`, data)
      toast.success('Статус обновлён! ✅')
      setSelectedIncident(null)
      statusForm.reset()
      loadData()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка обновления')
    }
  }

  const filteredIncidents = incidents.filter(inc => {
    // Фильтр по статусу
    if (filterStatus === 'active' && inc.status === 'closed') return false
    if (filterStatus === 'closed' && inc.status !== 'closed') return false
    
    // Фильтр по серьёзности
    if (filterSeverity !== 'all' && inc.severity !== filterSeverity) return false
    
    return true
  })

  // Сортировка: сначала high, потом по дате
  const sortedIncidents = [...filteredIncidents].sort((a, b) => {
    const severityOrder = { high: 0, medium: 1, low: 2 }
    if (severityOrder[a.severity] !== severityOrder[b.severity]) {
      return severityOrder[a.severity] - severityOrder[b.severity]
    }
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  })

  const stats = {
    total: incidents.length,
    open: incidents.filter(i => i.status === 'open').length,
    inProgress: incidents.filter(i => i.status === 'in_progress').length,
    closed: incidents.filter(i => i.status === 'closed').length,
    high: incidents.filter(i => i.severity === 'high' && i.status !== 'closed').length,
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
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <AlertTriangle className="w-8 h-8 text-orange-500" />
            Инциденты
          </h1>
          <p className="text-gray-500 mt-1">
            Регистрация и отслеживание инцидентов в аквапарке
          </p>
        </div>
        <Button onClick={() => {
          createForm.reset({ severity: 'medium' })
          setShowCreateModal(true)
        }}>
          <Plus className="w-4 h-4 mr-2" />
          Зарегистрировать инцидент
        </Button>
      </div>

      {/* Статистика */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <Card>
          <CardContent className="pt-4">
            <p className="text-sm text-gray-500">Всего</p>
            <p className="text-2xl font-bold">{stats.total}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <p className="text-sm text-gray-500">Открытые</p>
            <p className="text-2xl font-bold text-blue-600">{stats.open}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <p className="text-sm text-gray-500">В работе</p>
            <p className="text-2xl font-bold text-yellow-600">{stats.inProgress}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <p className="text-sm text-gray-500">Закрытые</p>
            <p className="text-2xl font-bold text-green-600">{stats.closed}</p>
          </CardContent>
        </Card>
        <Card className={stats.high > 0 ? 'border-red-500 bg-red-50' : ''}>
          <CardContent className="pt-4">
            <p className="text-sm text-gray-500">Высокая серьёзность</p>
            <p className="text-2xl font-bold text-red-600">{stats.high}</p>
          </CardContent>
        </Card>
      </div>

      {/* Фильтры */}
      <Card className="mb-6">
        <CardContent className="py-4">
          <div className="flex items-center gap-2 mb-3">
            <Filter className="w-4 h-4 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">Фильтры:</span>
          </div>
          <div className="flex gap-2 flex-wrap">
            <div className="flex gap-1 bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setFilterStatus('active')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterStatus === 'active' ? 'bg-white text-sky-600 shadow' : 'text-gray-600'
                )}
              >
                Активные
              </button>
              <button
                onClick={() => setFilterStatus('closed')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterStatus === 'closed' ? 'bg-white text-sky-600 shadow' : 'text-gray-600'
                )}
              >
                Закрытые
              </button>
              <button
                onClick={() => setFilterStatus('all')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterStatus === 'all' ? 'bg-white text-sky-600 shadow' : 'text-gray-600'
                )}
              >
                Все
              </button>
            </div>

            <div className="flex gap-1 bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setFilterSeverity('all')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterSeverity === 'all' ? 'bg-white text-sky-600 shadow' : 'text-gray-600'
                )}
              >
                Все серьёзности
              </button>
              <button
                onClick={() => setFilterSeverity('high')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterSeverity === 'high' ? 'bg-white text-red-600 shadow' : 'text-gray-600'
                )}
              >
                Высокая
              </button>
              <button
                onClick={() => setFilterSeverity('medium')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterSeverity === 'medium' ? 'bg-white text-yellow-600 shadow' : 'text-gray-600'
                )}
              >
                Средняя
              </button>
              <button
                onClick={() => setFilterSeverity('low')}
                className={clsx(
                  'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                  filterSeverity === 'low' ? 'bg-white text-blue-600 shadow' : 'text-gray-600'
                )}
              >
                Низкая
              </button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Список инцидентов */}
      {sortedIncidents.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <AlertTriangle className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">Инциденты не найдены</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {sortedIncidents.map(inc => {
            const sev = severityConfig[inc.severity]
            const stat = statusConfig[inc.status]
            const SevIcon = sev.icon
            const StatIcon = stat.icon

            return (
              <Card 
                key={inc.id} 
                className={clsx(
                  'hover:shadow-md transition-shadow',
                  inc.severity === 'high' && inc.status !== 'closed' && 'border-red-300'
                )}
              >
                <CardContent className="p-6">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className={clsx('px-2 py-1 rounded text-xs font-bold', sev.color)}>
                          <SevIcon className="w-3 h-3 inline mr-1" />
                          {sev.label}
                        </span>
                        <span className={clsx('px-2 py-1 rounded text-xs font-bold', stat.color)}>
                          <StatIcon className="w-3 h-3 inline mr-1" />
                          {stat.label}
                        </span>
                        {inc.severity === 'high' && inc.status !== 'closed' && (
                          <span className="px-2 py-1 rounded text-xs font-bold bg-red-500 text-white animate-pulse">
                            СРОЧНО
                          </span>
                        )}
                      </div>

                      <p className="text-base font-medium mb-2">{inc.description}</p>

                      <div className="flex flex-wrap gap-4 text-sm text-gray-500">
                        <div className="flex items-center gap-1">
                          <span className="font-medium">Зона:</span>
                          {inc.zone?.name || '—'}
                        </div>
                        <div className="flex items-center gap-1">
                          <span className="font-medium">Спасатель:</span>
                          {inc.lifeguard?.full_name || '—'}
                        </div>
                        <div className="flex items-center gap-1">
                          <span className="font-medium">Создан:</span>
                          {format(new Date(inc.created_at), 'dd MMM yyyy, HH:mm', { locale: ru })}
                        </div>
                      </div>

                      {inc.resolution && (
                        <div className="mt-3 p-3 bg-green-50 border border-green-200 rounded-lg">
                          <p className="text-xs font-semibold text-green-700 mb-1">Решение:</p>
                          <p className="text-sm text-green-800">{inc.resolution}</p>
                        </div>
                      )}
                    </div>

                    {inc.status !== 'closed' && (
                      <Button
                        size="sm"
                        onClick={() => {
                          setSelectedIncident(inc)
                          statusForm.reset({ status: inc.status })
                        }}
                      >
                        Изменить статус
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Модалка создания инцидента */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full">
            <div className="flex items-center justify-between p-6 border-b">
              <div className="flex items-center gap-3">
                <AlertTriangle className="w-6 h-6 text-orange-500" />
                <h2 className="text-xl font-bold">Регистрация инцидента</h2>
              </div>
              <button 
                onClick={() => {
                  setShowCreateModal(false)
                  createForm.reset()
                }} 
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={createForm.handleSubmit(onCreateSubmit)} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">Зона</label>
                <select
                  {...createForm.register('zone_id')}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
                >
                  <option value="">-- Выберите зону --</option>
                  {zones.map(zone => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
                {createForm.formState.errors.zone_id && (
                  <p className="text-sm text-red-500 mt-1">
                    {createForm.formState.errors.zone_id.message}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Серьёзность</label>
                <div className="grid grid-cols-3 gap-2">
                  {(['low', 'medium', 'high'] as const).map(level => {
                    const cfg = severityConfig[level]
                    const Icon = cfg.icon
                    const isSelected = createForm.watch('severity') === level

                    return (
                      <button
                        key={level}
                        type="button"
                        onClick={() => createForm.setValue('severity', level)}
                        className={clsx(
                          'p-3 border-2 rounded-lg text-center transition-all',
                          isSelected ? cfg.color + ' border-current' : 'border-gray-200 hover:border-gray-300'
                        )}
                      >
                        <Icon className="w-5 h-5 mx-auto mb-1" />
                        <div className="text-sm font-medium">{cfg.label}</div>
                      </button>
                    )
                  })}
                </div>
                <input type="hidden" {...createForm.register('severity')} />
                {createForm.formState.errors.severity && (
                  <p className="text-sm text-red-500 mt-1">
                    {createForm.formState.errors.severity.message}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Описание</label>
                <textarea
                  {...createForm.register('description')}
                  rows={4}
                  placeholder="Опишите что произошло..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
                />
                {createForm.formState.errors.description && (
                  <p className="text-sm text-red-500 mt-1">
                    {createForm.formState.errors.description.message}
                  </p>
                )}
              </div>

              <div className="flex gap-3 pt-4">
                <Button 
                  type="button" 
                  variant="secondary" 
                  onClick={() => {
                    setShowCreateModal(false)
                    createForm.reset()
                  }} 
                  className="flex-1"
                >
                  Отмена
                </Button>
                <Button type="submit" className="flex-1">
                   Зарегистрировать
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Модалка изменения статуса */}
      {selectedIncident && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full">
            <div className="flex items-center justify-between p-6 border-b">
              <div className="flex items-center gap-3">
                <Clock className="w-6 h-6 text-sky-500" />
                <h2 className="text-xl font-bold">Изменение статуса</h2>
              </div>
              <button 
                onClick={() => {
                  setSelectedIncident(null)
                  statusForm.reset()
                }} 
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={statusForm.handleSubmit(onStatusSubmit)} className="p-6 space-y-4">
              <div className="p-3 bg-gray-50 rounded-lg">
                <p className="text-sm text-gray-600">Инцидент:</p>
                <p className="font-medium">{selectedIncident.description}</p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Новый статус</label>
                <div className="grid grid-cols-3 gap-2">
                  {(['open', 'in_progress', 'closed'] as const).map(status => {
                    const cfg = statusConfig[status]
                    const Icon = cfg.icon
                    const isSelected = statusForm.watch('status') === status

                    return (
                      <button
                        key={status}
                        type="button"
                        onClick={() => statusForm.setValue('status', status)}
                        className={clsx(
                          'p-3 border-2 rounded-lg text-center transition-all',
                          isSelected ? cfg.color + ' border-current' : 'border-gray-200 hover:border-gray-300'
                        )}
                      >
                        <Icon className="w-5 h-5 mx-auto mb-1" />
                        <div className="text-sm font-medium">{cfg.label}</div>
                      </button>
                    )
                  })}
                </div>
                <input type="hidden" {...statusForm.register('status')} />
              </div>

              {statusForm.watch('status') === 'closed' && (
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Описание решения <span className="text-red-500">*</span>
                  </label>
                  <textarea
                    {...statusForm.register('resolution')}
                    rows={3}
                    placeholder="Как был решён инцидент..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                  />
                  {statusForm.formState.errors.resolution && (
                    <p className="text-sm text-red-500 mt-1">
                      {statusForm.formState.errors.resolution.message}
                    </p>
                  )}
                </div>
              )}

              <div className="flex gap-3 pt-4">
                <Button 
                  type="button" 
                  variant="secondary" 
                  onClick={() => {
                    setSelectedIncident(null)
                    statusForm.reset()
                  }} 
                  className="flex-1"
                >
                  Отмена
                </Button>
                <Button type="submit" className="flex-1">
                  ✅ Сохранить
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}