import { Outlet } from 'react-router';
import './i18n';
import { MainLayout } from '@/components/layout/MainLayout';

function App() {
  return (
    <MainLayout>
      <Outlet />
    </MainLayout>
  );
}

export default App
