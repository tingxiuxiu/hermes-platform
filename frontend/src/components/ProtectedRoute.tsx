import { Navigate, useLocation } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";

interface ProtectedRouteProps {
  children: React.ReactNode;
}

interface LocationState {
  from?: {
    pathname: string;
  };
}

function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated) {
    const state: LocationState = {
      from: {
        pathname: location.pathname,
      },
    };
    return <Navigate to="/login" state={state} replace />;
  }

  return <>{children}</>;
}

export default ProtectedRoute;
