import { StrictMode, lazy, Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './index.css'
import './i18n/index.ts'
import App from './App.tsx'
import ProtectedRoute from './components/ProtectedRoute.tsx'
import { useAuthStore } from './stores/authStore.ts'

const Dashboard = lazy(() => import('./pages/Dashboard.tsx'))
const Home = lazy(() => import('./pages/Home.tsx'))
const Login = lazy(() => import('./pages/Login.tsx'))
const Register = lazy(() => import('./pages/Register.tsx'))
const Users = lazy(() => import('./pages/admin/Users.tsx'))
const ExecutionDetail = lazy(() => import('./pages/ExecutionDetail.tsx'))
const Settings = lazy(() => import('./pages/Settings.tsx'))
const Profile = lazy(() => import('./pages/Profile.tsx'))
const ForgotPassword = lazy(() => import('./pages/ForgotPassword.tsx'))
const ProjectList = lazy(() => import('./pages/project/ProjectList.tsx'))
const VersionList = lazy(() => import('./pages/project/VersionList.tsx'))
const TestPlanList = lazy(() => import('./pages/project/TestPlanList.tsx'))
const TestCaseList = lazy(() => import('./pages/project/TestCaseList.tsx'))
const TestPlansIndex = lazy(() => import('./pages/test-plans/Index.tsx'))
const TestCasesIndex = lazy(() => import('./pages/test-cases/Index.tsx'))

function RootRedirect() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }
  
  return <Navigate to="/login" replace />;
}

function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="flex flex-col items-center gap-4">
        <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
        <span className="text-slate-500">Loading...</span>
      </div>
    </div>
  )
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      refetchOnWindowFocus: false,
    },
  },
})

const router = createBrowserRouter([
  {
    path: '/',
    element: <RootRedirect />,
  },
  {
    path: '/login',
    element: (
      <Suspense fallback={<PageLoader />}>
        <Login />
      </Suspense>
    ),
  },
  {
    path: '/register',
    element: (
      <Suspense fallback={<PageLoader />}>
        <Register />
      </Suspense>
    ),
  },
  {
    path: '/forgot-password',
    element: (
      <Suspense fallback={<PageLoader />}>
        <ForgotPassword />
      </Suspense>
    ),
  },
  {
    element: <App />,
    children: [
      {
        path: '/dashboard',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <Dashboard />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <Dashboard />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/tasks',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <Home />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/tasks/:id',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <ExecutionDetail />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/admin/users',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <Users />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/execution/:id',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <ExecutionDetail />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/settings',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <Settings />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/profile',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <Profile />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/projects',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <ProjectList />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/projects/:projectId/versions',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <VersionList />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/test-plans',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <TestPlansIndex />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/test-cases',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <TestCasesIndex />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/projects/:projectId/versions/:versionId/test-plans',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <TestPlanList />
            </Suspense>
          </ProtectedRoute>
        ),
      },
      {
        path: '/projects/:projectId/versions/:versionId/test-plans/:planId/test-cases',
        element: (
          <ProtectedRoute>
            <Suspense fallback={<PageLoader />}>
              <TestCaseList />
            </Suspense>
          </ProtectedRoute>
        ),
      },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
