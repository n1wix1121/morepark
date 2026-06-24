import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { ExportExcelButton } from '@/components/ui/ExportExcelButton'
import { useAuthStore } from '@/store/authStore'
import api from '@/api/client'
import { Users, MapPin, AlertTriangle } from 'lucide-react'
export function Dashboard() {
  const { user } = useAuthStore()
  const [zones, setZones] = useState<any[]>([])
  const [activeIncidents, setActiveIncidents] = useState(0)

  useEffect(() => {
    loadZones()
    loadActiveIncidents()
  }, [])

  const loadZones = async () => {
    try {
      const res = await api.get('/zones')
      setZones(res.data.zones || [])
    } catch (err) {
      console.error(err)
    }
  }

  const loadActiveIncidents = async () => {
    try {
      const res = await api.get('/incidents/active')
      setActiveIncidents(res.data.count ?? res.data.incidents?.length ?? 0)
    } catch {
      // Недоступно для ролей без доступа к инцидентам
    }
  }

  const totalVisitors = zones.reduce((sum, z) => sum + z.current_count, 0)
  const totalCapacity = zones.reduce((sum, z) => sum + z.capacity, 0)
  const loadPercent = totalCapacity > 0 ? (totalVisitors / totalCapacity) * 100 : 0

  return (
    <div>
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold">Добро пожаловать, {user?.full_name}!</h1>
          <p className="text-gray-500 mt-1">Панель управления аквапарком «Море Парк»</p>
        </div>
        {user?.role === 'director' && (
          <ExportExcelButton
            endpoint="/reports/summary/excel"
            filename="svodka_buhgalteriya.xlsx"
            label="Сводка для бухгалтерии"
          />
        )}
      </div>

      {/* KPI карточки */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Посетителей сейчас</CardTitle>
            <Users className="w-4 h-4 text-sky-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{totalVisitors}</div>
            <p className="text-xs text-gray-500 mt-1">из {totalCapacity} мест</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Загрузка</CardTitle>
            <MapPin className="w-4 h-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{loadPercent.toFixed(0)}%</div>
            <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
              <div 
                className="bg-sky-500 h-2 rounded-full transition-all"
                style={{ width: `${Math.min(loadPercent, 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Активных зон</CardTitle>
            <MapPin className="w-4 h-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{zones.length}</div>
            <p className="text-xs text-gray-500 mt-1">работают сейчас</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Инциденты</CardTitle>
            <AlertTriangle className="w-4 h-4 text-orange-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{activeIncidents}</div>
            <p className="text-xs text-gray-500 mt-1">открытых</p>
          </CardContent>
        </Card>
      </div>

      {/* Карта зон */}
      <Card>
        <CardHeader>
          <CardTitle>Загрузка зон в реальном времени</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {zones.map(zone => {
              const percent = (zone.current_count / zone.capacity) * 100
              const isFull = percent >= 100
              const isBusy = percent >= 80

              return (
                <div 
                  key={zone.id}
                  className="p-4 border rounded-lg"
                >
                  <div className="flex justify-between items-start mb-2">
                    <h3 className="font-semibold">{zone.name}</h3>
                    {isFull && <span className="text-xs bg-red-100 text-red-600 px-2 py-1 rounded">ЗАПОЛНЕНА</span>}
                    {isBusy && !isFull && <span className="text-xs bg-yellow-100 text-yellow-600 px-2 py-1 rounded">ЗАГРУЖЕНА</span>}
                  </div>
                  <p className="text-2xl font-bold mb-2">
                    {zone.current_count} <span className="text-sm text-gray-500">/ {zone.capacity}</span>
                  </p>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className={`h-2 rounded-full transition-all ${
                        isFull ? 'bg-red-500' : isBusy ? 'bg-yellow-500' : 'bg-green-500'
                      }`}
                      style={{ width: `${Math.min(percent, 100)}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}