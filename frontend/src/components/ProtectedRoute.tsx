import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, isInitialized } = useAuth();

  // Wait for auth to initialize from localStorage
  if (!isInitialized) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  // Now check if authenticated
  if (!token) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}