import { useEffect, useState } from 'react'
import { useForm, SubmitHandler } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { MapPin, Plus, Edit2, Trash2, Users, X } from 'lucide-react'
import { clsx } from 'clsx'

const zoneSchema = z.object({
  name: z.string().min(2, 'Название минимум 2 символа'),
  description: z.string().optional(),
  capacity: z.coerce.number().min(1, 'Вместимость минимум 1'),
})

type ZoneData = z.infer<typeof zoneSchema>

interface Zone {
  id: string
  name: string
  description: string
  capacity: number
  current_count: number
  created_at: string
}

export function ZonesList() {
  const [zones, setZones] = useState<Zone[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingZone, setEditingZone] = useState<Zone | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ZoneData>({
    resolver: zodResolver(zoneSchema),
  })

  useEffect(() => {
    loadZones()
  }, [])

  const loadZones = async () => {
    setLoading(true)
    try {
      const res = await api.get('/zones')
      setZones(res.data.zones || [])
    } catch (err) {
      console.error(err)
      toast.error('Ошибка загрузки зон')
    } finally {
      setLoading(false)
    }
  }

  const onSubmit: SubmitHandler<ZoneData> = async (data) => {
    try {
      if (editingZone) {
        await api.put(`/zones/${editingZone.id}`, data)
        toast.success('Зона обновлена! ✅')
      } else {
        await api.post('/zones', data)
        toast.success('Зона создана! ✅')
      }
      setShowModal(false)
      setEditingZone(null)
      reset()
      loadZones()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка сохранения')
    }
  }

  const handleEdit = (zone: Zone) => {
    setEditingZone(zone)
    reset({
      name: zone.name,
      description: zone.description,
      capacity: zone.capacity,
    })
    setShowModal(true)
  }

  const handleDelete = async (zoneId: string) => {
    if (!confirm('Вы уверены, что хотите удалить эту зону?')) return

    try {
      await api.delete(`/zones/${zoneId}`)
      toast.success('Зона удалена! ️')
      loadZones()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка удаления')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Загрузка...</p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <MapPin className="w-8 h-8 text-sky-500" />
            Управление зонами
          </h1>
          <p className="text-gray-500 mt-1">
            Добавление и редактирование зон аквапарка
          </p>
        </div>
        <Button onClick={() => {
          setEditingZone(null)
          reset()
          setShowModal(true)
        }}>
          <Plus className="w-4 h-4 mr-2" />
          Добавить зону
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {zones.map(zone => (
          <Card key={zone.id} className="hover:shadow-md transition-shadow">
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <CardTitle className="text-lg">{zone.name}</CardTitle>
                  <p className="text-sm text-gray-500 mt-1 line-clamp-2">
                    {zone.description || 'Нет описания'}
                  </p>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-gray-400" />
                <span className="text-sm">
                  <strong>{zone.current_count}</strong> / {zone.capacity} посетителей
                </span>
              </div>

              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className={clsx(
                    'h-2 rounded-full transition-all',
                    zone.current_count >= zone.capacity ? 'bg-red-500' :
                    zone.current_count >= zone.capacity * 0.8 ? 'bg-yellow-500' : 'bg-green-500'
                  )}
                  style={{ width: `${Math.min((zone.current_count / zone.capacity) * 100, 100)}%` }}
                />
              </div>

              <div className="flex gap-2 pt-2">
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => handleEdit(zone)}
                  className="flex-1"
                >
                  <Edit2 className="w-4 h-4 mr-1" />
                  Редактировать
                </Button>
                <Button
                  size="sm"
                  variant="danger"
                  onClick={() => handleDelete(zone.id)}
                  disabled={zone.current_count > 0}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {zones.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <MapPin className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">Зоны не добавлены</p>
          </CardContent>
        </Card>
      )}

      {/* Модалка создания/редактирования */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
            <div className="flex items-center justify-between p-6 border-b">
              <div className="flex items-center gap-3">
                <MapPin className="w-6 h-6 text-sky-500" />
                <h2 className="text-xl font-bold">
                  {editingZone ? 'Редактирование зоны' : 'Новая зона'}
                </h2>
              </div>
              <button 
                onClick={() => {
                  setShowModal(false)
                  setEditingZone(null)
                  reset()
                }} 
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="p-6 space-y-4">
              <Input
                label="Название зоны"
                placeholder="Волновой бассейн"
                error={errors.name?.message}
                {...register('name')}
              />

              <div>
                <label className="block text-sm font-medium mb-2">Описание</label>
                <textarea
                  {...register('description')}
                  rows={3}
                  placeholder="Описание зоны..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                />
              </div>

              <Input
                label="Вместимость (человек)"
                type="number"
                placeholder="50"
                error={errors.capacity?.message}
                {...register('capacity')}
              />

              <div className="flex gap-3 pt-4">
                <Button 
                  type="button" 
                  variant="secondary" 
                  onClick={() => {
                    setShowModal(false)
                    setEditingZone(null)
                    reset()
                  }} 
                  className="flex-1"
                >
                  Отмена
                </Button>
                <Button type="submit" className="flex-1">
                  {editingZone ? ' Сохранить' : '✅ Создать'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}