import { useEffect, useState } from 'react'
import api from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { ExportExcelButton } from '@/components/ui/ExportExcelButton'
import { History, Ticket, DollarSign } from 'lucide-react'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'

interface Sale {
  id: string
  amount: number
  payment_method: string
  is_refund: boolean
  created_at: string
  cashier: {
    full_name: string
  }
  ticket: {
    zone: {
      name: string
    }
  }
}

export function SalesHistory() {
  const [sales, setSales] = useState<Sale[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadSales()
  }, [])

  const loadSales = async () => {
    try {
      const res = await api.get('/tickets?limit=100')
      setSales(res.data.sales || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const totalRevenue = sales
    .filter(s => !s.is_refund)
    .reduce((sum, s) => sum + s.amount, 0)

  const totalRefunds = sales
    .filter(s => s.is_refund)
    .reduce((sum, s) => sum + Math.abs(s.amount), 0)

  return (
    <div>
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <History className="w-8 h-8 text-sky-500" />
            История продаж
          </h1>
          <p className="text-gray-500 mt-1">
            Все операции по продаже и возврату билетов
          </p>
        </div>
        <ExportExcelButton
          endpoint="/reports/sales/excel"
          filename="prodazhi.xlsx"
          label="Excel для бухгалтерии"
        />
      </div>

      {/* KPI */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-green-100 rounded-lg">
                <DollarSign className="w-6 h-6 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Выручка</p>
                <p className="text-2xl font-bold text-green-600">
                  {totalRevenue.toLocaleString()} ₽
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-sky-100 rounded-lg">
                <Ticket className="w-6 h-6 text-sky-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Продано билетов</p>
                <p className="text-2xl font-bold">
                  {sales.filter(s => !s.is_refund).length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-red-100 rounded-lg">
                <DollarSign className="w-6 h-6 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Возвратов</p>
                <p className="text-2xl font-bold text-red-600">
                  {totalRefunds.toLocaleString()} ₽
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Таблица продаж */}
      <Card>
        <CardHeader>
          <CardTitle>Последние операции</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p className="text-center py-8 text-gray-500">Загрузка...</p>
          ) : sales.length === 0 ? (
            <p className="text-center py-8 text-gray-500">
              Продаж пока нет
            </p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b text-left text-sm text-gray-500">
                    <th className="pb-3">Дата</th>
                    <th className="pb-3">Зона</th>
                    <th className="pb-3">Кассир</th>
                    <th className="pb-3">Оплата</th>
                    <th className="pb-3 text-right">Сумма</th>
                    <th className="pb-3">Статус</th>
                  </tr>
                </thead>
                <tbody>
                  {sales.map(sale => (
                    <tr key={sale.id} className="border-b hover:bg-gray-50">
                      <td className="py-3 text-sm">
                        {format(new Date(sale.created_at), 'dd MMM yyyy, HH:mm', {
                          locale: ru,
                        })}
                      </td>
                      <td className="py-3 text-sm">
                        {sale.ticket?.zone?.name || '—'}
                      </td>
                      <td className="py-3 text-sm">
                        {sale.cashier?.full_name || 'Онлайн'}
                      </td>
                      <td className="py-3 text-sm">
                        {sale.payment_method === 'cash'
                          ? '💵 Наличные'
                          : sale.payment_method === 'card'
                          ? '💳 Карта'
                          : sale.payment_method}
                      </td>
                      <td className="py-3 text-right font-semibold">
                        <span
                          className={
                            sale.is_refund ? 'text-red-600' : 'text-green-600'
                          }
                        >
                          {sale.is_refund ? '-' : '+'}
                          {Math.abs(sale.amount).toLocaleString()} ₽
                        </span>
                      </td>
                      <td className="py-3">
                        <span
                          className={`text-xs px-2 py-1 rounded ${
                            sale.is_refund
                              ? 'bg-red-100 text-red-700'
                              : 'bg-green-100 text-green-700'
                          }`}
                        >
                          {sale.is_refund ? 'Возврат' : 'Продажа'}
                        </span>
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
  )
}