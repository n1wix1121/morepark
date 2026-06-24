import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from 'sonner'
import { Login } from '@/pages/Login'
import { Dashboard } from '@/pages/admin/Dashboard'
import { SellTicket } from '@/pages/admin/Tickets/SellTicket'
import { SalesHistory } from '@/pages/admin/Tickets/SalesHistory'
import { WaterQuality } from '@/pages/admin/Water/WaterQuality'
import { EquipmentList } from '@/pages/admin/Equipment/EquipmentList'
import { InventoryList } from '@/pages/admin/Inventory/InventoryList'
import { IncidentsList } from '@/pages/admin/Incidents/IncidentsList'
import { ZonesList } from '@/pages/admin/Zones/ZonesList'
import { UsersList } from '@/pages/admin/Users/UsersList'
import { BuyTicket } from '@/pages/public/BuyTicket'
import { AdminLayout } from '@/components/layout/AdminLayout'
import { ProtectedRoute } from '@/components/layout/ProtectedRoute'

function App() {
  return (
    <BrowserRouter>
      <Toaster position="top-right" richColors />
      <Routes>
        <Route path="/buy" element={<BuyTicket />} />
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Navigate to="/login" replace />} />

        <Route
          path="/admin"
          element={
            <ProtectedRoute>
              <AdminLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          
          <Route
            path="tickets"
            element={
              <ProtectedRoute roles={['cashier', 'director']}>
                <SellTicket />
              </ProtectedRoute>
            }
          />
          <Route
            path="tickets/history"
            element={
              <ProtectedRoute roles={['cashier', 'director']}>
                <SalesHistory />
              </ProtectedRoute>
            }
          />
          
          <Route
            path="water"
            element={
              <ProtectedRoute roles={['technician', 'director']}>
                <WaterQuality />
              </ProtectedRoute>
            }
          />
          
          <Route
            path="equipment"
            element={
              <ProtectedRoute roles={['technician', 'director']}>
                <EquipmentList />
              </ProtectedRoute>
            }
          />
          
          <Route
            path="inventory"
            element={
              <ProtectedRoute roles={['director', 'barman']}>
                <InventoryList />
              </ProtectedRoute>
            }
          />
          
          <Route
            path="incidents"
            element={
              <ProtectedRoute roles={['lifeguard', 'director']}>
                <IncidentsList />
              </ProtectedRoute>
            }
          />
          
          {/*  НОВЫЕ МАРШРУТЫ */}
          <Route
            path="zones"
            element={
              <ProtectedRoute roles={['director']}>
                <ZonesList />
              </ProtectedRoute>
            }
          />
          
          <Route
            path="users"
            element={
              <ProtectedRoute roles={['director']}>
                <UsersList />
              </ProtectedRoute>
            }
          />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App