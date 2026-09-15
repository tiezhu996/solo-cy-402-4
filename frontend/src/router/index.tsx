import { createBrowserRouter, Navigate } from 'react-router-dom'
import Layout from '@/pages/Layout'
import Login from '@/pages/Login'
import Cases from '@/pages/Cases'
import CaseDetail from '@/pages/CaseDetail'
import Clients from '@/pages/Clients'
import Billing from '@/pages/Billing'
import Documents from '@/pages/Documents'
import Profile from '@/pages/Profile'
import AuditLogs from '@/pages/AuditLogs'
import { RequireAuth, RequireRole } from './guards'

const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: (
      <RequireAuth>
        <Layout />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <Navigate to="/cases" replace /> },
      { path: 'cases', element: <Cases /> },
      { path: 'cases/:id', element: <CaseDetail /> },
      { path: 'clients', element: <Clients /> },
      { path: 'billing', element: <Billing /> },
      { path: 'documents', element: <Documents /> },
      { path: 'profile', element: <Profile /> },
      {
        path: 'audit-logs',
        element: (
          <RequireRole roles={['admin']}>
            <AuditLogs />
          </RequireRole>
        ),
      },
    ],
  },
])

export default router
