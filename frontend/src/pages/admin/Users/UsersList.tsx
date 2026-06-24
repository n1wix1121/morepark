import { useEffect, useState } from 'react'
import { useForm, SubmitHandler } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import api from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Users, Plus, Edit2, Trash2, X, Shield, UserCheck, UserX } from 'lucide-react'
import { clsx } from 'clsx'

const userSchema = z.object({
  email: z.string().email('Неверный email'),
  full_name: z.string().min(2, 'Имя минимум 2 символа'),
  password: z.string().min(6, 'Пароль минимум 6 символов').optional(),
  role: z.enum(['director', 'cashier', 'lifeguard', 'technician', 'barman']),
  is_active: z.boolean(),
})

type UserData = z.infer<typeof userSchema>

interface User {
  id: string
  email: string
  full_name: string
  role: string
  is_active: boolean
  created_at: string
}

const roleLabels: Record<string, string> = {
  director: 'Директор',
  cashier: 'Кассир',
  lifeguard: 'Спасатель',
  technician: 'Тех. служба',
  barman: 'Бармен',
}

const roleColors: Record<string, string> = {
  director: 'bg-purple-100 text-purple-700',
  cashier: 'bg-blue-100 text-blue-700',
  lifeguard: 'bg-orange-100 text-orange-700',
  technician: 'bg-green-100 text-green-700',
  barman: 'bg-pink-100 text-pink-700',
}

export function UsersList() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<UserData>({
    resolver: zodResolver(userSchema),
  })

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    setLoading(true)
    try {
      const res = await api.get('/users')
      setUsers(res.data.users || [])
    } catch (err) {
      console.error(err)
      toast.error('Ошибка загрузки пользователей')
    } finally {
      setLoading(false)
    }
  }

  const onSubmit: SubmitHandler<UserData> = async (data) => {
    try {
      if (editingUser) {
        await api.put(`/users/${editingUser.id}`, data)
        toast.success('Пользователь обновлён! ✅')
      } else {
        await api.post('/users', data)
        toast.success('Пользователь создан! ✅')
      }
      setShowModal(false)
      setEditingUser(null)
      reset()
      loadUsers()
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Ошибка сохранения')
    }
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    reset({
      email: user.email,
      full_name: user.full_name,
      role: user.role as any,
      is_active: user.is_active,
    })
    setShowModal(true)
  }

  const handleDelete = async (userId: string) => {
    if (!confirm('Вы уверены, что хотите удалить этого пользователя?')) return

    try {
      await api.delete(`/users/${userId}`)
      toast.success('Пользователь удалён! ️')
      loadUsers()
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
            <Users className="w-8 h-8 text-sky-500" />
            Управление пользователями
          </h1>
          <p className="text-gray-500 mt-1">
            Добавление и редактирование сотрудников
          </p>
        </div>
        <Button onClick={() => {
          setEditingUser(null)
          reset({ is_active: true })
          setShowModal(true)
        }}>
          <Plus className="w-4 h-4 mr-2" />
          Добавить пользователя
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b text-left text-sm text-gray-500">
                  <th className="p-4">Имя</th>
                  <th className="p-4">Email</th>
                  <th className="p-4">Роль</th>
                  <th className="p-4">Статус</th>
                  <th className="p-4 text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {users.map(user => (
                  <tr key={user.id} className="border-b hover:bg-gray-50">
                    <td className="p-4">
                      <div className="font-medium">{user.full_name}</div>
                    </td>
                    <td className="p-4 text-sm text-gray-600">{user.email}</td>
                    <td className="p-4">
                      <span className={clsx('px-2 py-1 rounded text-xs font-semibold', roleColors[user.role])}>
                        {roleLabels[user.role]}
                      </span>
                    </td>
                    <td className="p-4">
                      {user.is_active ? (
                        <span className="flex items-center gap-1 text-xs text-green-700">
                          <UserCheck className="w-4 h-4" /> Активен
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-xs text-red-700">
                          <UserX className="w-4 h-4" /> Неактивен
                        </span>
                      )}
                    </td>
                    <td className="p-4 text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() => handleEdit(user)}
                        >
                          <Edit2 className="w-4 h-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="danger"
                          onClick={() => handleDelete(user.id)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {users.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <Users className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">Пользователи не добавлены</p>
          </CardContent>
        </Card>
      )}

      {/* Модалка создания/редактирования */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
            <div className="flex items-center justify-between p-6 border-b">
              <div className="flex items-center gap-3">
                <Shield className="w-6 h-6 text-sky-500" />
                <h2 className="text-xl font-bold">
                  {editingUser ? 'Редактирование пользователя' : 'Новый пользователь'}
                </h2>
              </div>
              <button 
                onClick={() => {
                  setShowModal(false)
                  setEditingUser(null)
                  reset()
                }} 
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="p-6 space-y-4">
              <Input
                label="Email"
                type="email"
                placeholder="user@morepark.ru"
                error={errors.email?.message}
                {...register('email')}
              />

              <Input
                label="ФИО"
                placeholder="Иванов Иван Иванович"
                error={errors.full_name?.message}
                {...register('full_name')}
              />

              {!editingUser && (
                <Input
                  label="Пароль"
                  type="password"
                  placeholder="Минимум 6 символов"
                  error={errors.password?.message}
                  {...register('password')}
                />
              )}

              <div>
                <label className="block text-sm font-medium mb-2">Роль</label>
                <select
                  {...register('role')}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                >
                  {Object.entries(roleLabels).map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </select>
                {errors.role && (
                  <p className="text-sm text-red-500 mt-1">{errors.role.message}</p>
                )}
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  {...register('is_active')}
                  className="w-4 h-4 text-sky-600 rounded"
                />
                <label htmlFor="is_active" className="text-sm font-medium">
                  Активен
                </label>
              </div>

              <div className="flex gap-3 pt-4">
                <Button 
                  type="button" 
                  variant="secondary" 
                  onClick={() => {
                    setShowModal(false)
                    setEditingUser(null)
                    reset()
                  }} 
                  className="flex-1"
                >
                  Отмена
                </Button>
                <Button type="submit" className="flex-1">
                  {editingUser ? '💾 Сохранить' : '✅ Создать'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}