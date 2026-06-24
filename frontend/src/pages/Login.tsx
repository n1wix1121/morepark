import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { useAuthStore } from '@/store/authStore'
import { Waves } from 'lucide-react'

const loginSchema = z.object({
  email: z.string().email('Неверный email'),
  password: z.string().min(6, 'Пароль минимум 6 символов'),
})

type LoginData = z.infer<typeof loginSchema>

export function Login() {
  const navigate = useNavigate()
  const { login, isLoading } = useAuthStore()
  const [error, setError] = useState('')

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginData) => {
    try {
      setError('')
      await login(data.email, data.password)
      toast.success('Добро пожаловать!')
      navigate('/admin')
    } catch (err: any) {
      const message = err.response?.data?.error || 'Ошибка входа'
      setError(message)
      toast.error(message)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-sky-400 to-blue-600 p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="w-16 h-16 bg-sky-100 rounded-full flex items-center justify-center">
              <Waves className="w-8 h-8 text-sky-500" />
            </div>
          </div>
          <CardTitle className="text-2xl">Море Парк</CardTitle>
          <p className="text-sm text-gray-500 mt-1">
            Вход в систему управления аквапарком
          </p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <Input
              label="Email"
              type="email"
              placeholder="director@morepark.ru"
              error={errors.email?.message}
              {...register('email')}
            />
            <Input
              label="Пароль"
              type="password"
              placeholder="••••••"
              error={errors.password?.message}
              {...register('password')}
            />

            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-600">
                {error}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Вход...' : 'Войти'}
            </Button>
          </form>

          <div className="mt-6 p-4 bg-gray-50 rounded-md text-xs text-gray-600">
            <p className="font-semibold mb-2">Тестовые аккаунты:</p>
            <p>📧 director@morepark.ru / test123</p>
            <p>📧 cashier@morepark.ru / test123</p>
            <p>📧 lifeguard@morepark.ru / test123</p>
            <p>📧 technician@morepark.ru / test123</p>
            <p>📧 barman@morepark.ru / test123</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}