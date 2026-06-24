import { Link, useLocation } from 'react-router-dom'
import { 
  LayoutDashboard, Map, Droplets, Wrench, Package, 
  AlertTriangle, Users, Ticket, LogOut 
} from 'lucide-react'
import { useAuthStore } from '@/store/authStore'
import { clsx } from 'clsx'

const menuItems = [
  { path: '/admin', icon: LayoutDashboard, label: 'Дашборд', roles: ['director', 'cashier', 'lifeguard', 'technician', 'barman'] },
  { path: '/admin/zones', icon: Map, label: 'Зоны', roles: ['director'] },
  { path: '/admin/tickets', icon: Ticket, label: 'Продажа билетов', roles: ['director', 'cashier'] },
  { path: '/admin/tickets/history', icon: Ticket, label: 'История продаж', roles: ['director', 'cashier'] },
  { path: '/admin/water', icon: Droplets, label: 'Водоподготовка', roles: ['director', 'technician'] },
  { path: '/admin/equipment', icon: Wrench, label: 'Оборудование', roles: ['director', 'technician'] },
  { path: '/admin/inventory', icon: Package, label: 'Склад', roles: ['director', 'barman'] },
  { path: '/admin/incidents', icon: AlertTriangle, label: 'Инциденты', roles: ['director', 'lifeguard'] },
  { path: '/admin/users', icon: Users, label: 'Пользователи', roles: ['director'] },
]

export function Sidebar() {
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const filteredMenu = menuItems.filter(item => 
    user && item.roles.includes(user.role)
  )

  return (
    <aside className="w-64 bg-white border-r border-gray-200 flex flex-col">
      {/* Логотип */}
      <div className="p-6 border-b">
        <h1 className="text-2xl font-bold text-sky-500">🌊 Море Парк</h1>
        <p className="text-xs text-gray-500 mt-1">Система управления</p>
      </div>

      {/* Меню */}
      <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
        {filteredMenu.map(item => {
          const Icon = item.icon
          const isActive = location.pathname === item.path
          
          return (
            <Link
              key={item.path}
              to={item.path}
              className={clsx(
                'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                isActive 
                  ? 'bg-sky-50 text-sky-600' 
                  : 'text-gray-700 hover:bg-gray-100'
              )}
            >
              <Icon className="w-5 h-5" />
              {item.label}
            </Link>
          )
        })}
      </nav>

      {/* Пользователь */}
      <div className="p-4 border-t">
        <div className="mb-3">
          <p className="text-sm font-medium">{user?.full_name}</p>
          <p className="text-xs text-gray-500 capitalize">
            {user?.role === 'director' && 'Директор'}
            {user?.role === 'cashier' && 'Кассир'}
            {user?.role === 'lifeguard' && 'Спасатель'}
            {user?.role === 'technician' && 'Тех. служба'}
            {user?.role === 'barman' && 'Бармен'}
          </p>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-2 w-full px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-md"
        >
          <LogOut className="w-4 h-4" />
          Выйти
        </button>
      </div>
    </aside>
  )
}