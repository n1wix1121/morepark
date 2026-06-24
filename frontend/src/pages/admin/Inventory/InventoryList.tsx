import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { ExportExcelButton } from '@/components/ui/ExportExcelButton'
import { Package, AlertTriangle, TrendingUp, TrendingDown, Clock, CheckCircle } from 'lucide-react'
import { clsx } from 'clsx'
import { format, differenceInDays } from 'date-fns'
import { ru } from 'date-fns/locale'
import { MovementModal } from './MovementModal'

interface InventoryItem {
  id: string
  name: string
  category: string
  quantity: number
  unit: string
  min_quantity: number
  expiry_date: string | null
  price: number
  is_low_stock: boolean
  is_expiring_soon: boolean
  is_expired: boolean
}

interface Movement {
  id: string
  type: 'in' | 'out'
  quantity: number
  reason: string
  created_at: string
  user: { full_name: string }
}

const categoryLabels: Record<string, string> = {
  chemical: 'Химия',
  drinks: 'Напитки',
  food: 'Еда',
  supplies: 'Расходники',
}

const categoryColors: Record<string, string> = {
  chemical: 'bg-purple-100 text-purple-700',
  drinks: 'bg-blue-100 text-blue-700',
  food: 'bg-orange-100 text-orange-700',
  supplies: 'bg-gray-100 text-gray-700',
}

export function InventoryList() {
  const [items, setItems] = useState<InventoryItem[]>([])
  const [movements, setMovements] = useState<Movement[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedItem, setSelectedItem] = useState<{
    item: InventoryItem
    type: 'in' | 'out'
  } | null>(null)
  const [filter, setFilter] = useState<string>('all')

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await api.get('/inventory')
      setItems(res.data.inventory || [])
    } catch (err) {
      console.error(err)
      toast.error('Ошибка загрузки склада')
    } finally {
      setLoading(false)
    }
  }

  const loadMovements = async (itemId: string) => {
    try {
      const res = await api.get(`/inventory/${itemId}/movements?limit=10`)
      setMovements(res.data.movements || [])
    } catch (err) {
      console.error(err)
    }
  }

  const filteredItems = filter === 'all' 
    ? items 
    : filter === 'low' 
    ? items.filter(i => i.is_low_stock)
    : filter === 'expiring'
    ? items.filter(i => i.is_expiring_soon || i.is_expired)
    : items.filter(i => i.category === filter)

  const stats = {
    total: items.length,
    lowStock: items.filter(i => i.is_low_stock).length,
    expiring: items.filter(i => i.is_expiring_soon).length,
    expired: items.filter(i => i.is_expired).length,
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
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <Package className="w-8 h-8 text-sky-500" />
            Склад ТМЦ
          </h1>
          <p className="text-gray-500 mt-1">
            Учёт товаров, контроль остатков и сроков годности
          </p>
        </div>
        <ExportExcelButton
          endpoint="/reports/inventory/excel"
          filename="sklad.xlsx"
          label="Excel для бухгалтерии"
        />
      </div>

      {/* Статистика */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-sky-100 rounded-lg">
                <Package className="w-6 h-6 text-sky-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Всего позиций</p>
                <p className="text-2xl font-bold">{stats.total}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className={stats.lowStock > 0 ? 'border-yellow-500' : ''}>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-yellow-100 rounded-lg">
                <AlertTriangle className="w-6 h-6 text-yellow-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Низкий остаток</p>
                <p className="text-2xl font-bold text-yellow-600">{stats.lowStock}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className={stats.expiring > 0 ? 'border-orange-500' : ''}>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-orange-100 rounded-lg">
                <Clock className="w-6 h-6 text-orange-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Истекает срок</p>
                <p className="text-2xl font-bold text-orange-600">{stats.expiring}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className={stats.expired > 0 ? 'border-red-500' : ''}>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-red-100 rounded-lg">
                <AlertTriangle className="w-6 h-6 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Просрочено</p>
                <p className="text-2xl font-bold text-red-600">{stats.expired}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Фильтры */}
      <Card className="mb-6">
        <CardContent className="py-4">
          <div className="flex gap-2 flex-wrap">
            <button
              onClick={() => setFilter('all')}
              className={clsx(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                filter === 'all' ? 'bg-sky-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              )}
            >
              Все ({items.length})
            </button>
            <button
              onClick={() => setFilter('low')}
              className={clsx(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                filter === 'low' ? 'bg-yellow-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              )}
            >
              Низкий остаток ({stats.lowStock})
            </button>
            <button
              onClick={() => setFilter('expiring')}
              className={clsx(
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                filter === 'expiring' ? 'bg-orange-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              )}
            >
              Истекает срок ({stats.expiring + stats.expired})
            </button>
            {Object.entries(categoryLabels).map(([key, label]) => (
              <button
                key={key}
                onClick={() => setFilter(key)}
                className={clsx(
                  'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                  filter === key ? 'bg-sky-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                )}
              >
                {label}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Список товаров */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredItems.map(item => {
          const daysUntilExpiry = item.expiry_date 
            ? differenceInDays(new Date(item.expiry_date), new Date())
            : null

          return (
            <Card 
              key={item.id} 
              className={clsx(
                'hover:shadow-md transition-shadow',
                item.is_expired && 'border-red-500',
                item.is_low_stock && !item.is_expired && 'border-yellow-500',
                item.is_expiring_soon && !item.is_expired && 'border-orange-500'
              )}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <CardTitle className="text-lg">{item.name}</CardTitle>
                    <span className={clsx('inline-block mt-1 px-2 py-0.5 rounded text-xs font-semibold', categoryColors[item.category])}>
                      {categoryLabels[item.category]}
                    </span>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {/* Остаток */}
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-500">Остаток:</span>
                  <div className="text-right">
                    <span className={clsx(
                      'text-2xl font-bold',
                      item.is_low_stock ? 'text-red-600' : 'text-green-600'
                    )}>
                      {item.quantity}
                    </span>
                    <span className="text-sm text-gray-500 ml-1">{item.unit}</span>
                  </div>
                </div>

                {/* Прогресс-бар остатка */}
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={clsx(
                      'h-2 rounded-full transition-all',
                      item.quantity <= item.min_quantity * 0.5 ? 'bg-red-500' :
                      item.quantity <= item.min_quantity ? 'bg-yellow-500' : 'bg-green-500'
                    )}
                    style={{ 
                      width: `${Math.min((item.quantity / (item.min_quantity * 3)) * 100, 100)}%` 
                    }}
                  />
                </div>

                <div className="text-xs text-gray-500">
                  Минимум: {item.min_quantity} {item.unit}
                </div>

                {/* Срок годности */}
                {item.expiry_date && (
                  <div className={clsx(
                    'p-2 rounded-lg text-sm',
                    item.is_expired ? 'bg-red-50 text-red-700' :
                    item.is_expiring_soon ? 'bg-orange-50 text-orange-700' : 'bg-green-50 text-green-700'
                  )}>
                    <div className="flex items-center gap-2">
                      <Clock className="w-4 h-4" />
                      <span>
                        {item.is_expired 
                          ? `Просрочено ${Math.abs(daysUntilExpiry!)} дн.`
                          : `Годен до ${format(new Date(item.expiry_date), 'dd MMM yyyy', { locale: ru })}`
                        }
                      </span>
                    </div>
                  </div>
                )}

                {/* Кнопки действий */}
                <div className="flex gap-2 pt-2">
                  <Button
                    size="sm"
                    onClick={() => {
                      setSelectedItem({ item, type: 'in' })
                      loadMovements(item.id)
                    }}
                    className="flex-1"
                  >
                    <TrendingUp className="w-4 h-4 mr-1" />
                    Приход
                  </Button>
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => {
                      setSelectedItem({ item, type: 'out' })
                      loadMovements(item.id)
                    }}
                    className="flex-1"
                    disabled={item.quantity <= 0}
                  >
                    <TrendingDown className="w-4 h-4 mr-1" />
                    Расход
                  </Button>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {filteredItems.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <Package className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">Товары не найдены</p>
            <p className="text-sm text-gray-400 mt-2">Перезапустите backend — тестовые данные подгрузятся автоматически</p>
          </CardContent>
        </Card>
      )}

      {/* Модалка движения */}
      {selectedItem && (
        <MovementModal
          inventoryId={selectedItem.item.id}
          itemName={selectedItem.item.name}
          currentQuantity={selectedItem.item.quantity}
          type={selectedItem.type}
          onClose={() => setSelectedItem(null)}
          onSuccess={() => {
            loadData()
            loadMovements(selectedItem.item.id)
          }}
        />
      )}
    </div>
  )
}