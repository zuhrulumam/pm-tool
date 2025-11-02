import { GoogleLogin, GoogleOAuthProvider } from '@react-oauth/google';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../lib/api';
import toast from 'react-hot-toast';

function LoginContent() {
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSuccess = async (credentialResponse: { credential?: string }) => {
    try {
      const { data } = await api.post('/auth/google/callback', {
        credential: credentialResponse.credential
      });

      login(data.data.token, data.data.user);
      toast.success('Login successful!');
      navigate('/projects');
    } catch (error) {
      toast.error('Login failed. Please try again.');
      console.error('Login error:', error);
    }
  };

  const handleError = () => {
    toast.error('Login failed. Please try again.');
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center px-4">
      <div className="max-w-md w-full bg-white rounded-2xl shadow-lg p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Projects & Notes
          </h1>
          <p className="text-gray-600">
            Sign in to manage your projects and notes
          </p>
        </div>

        <div className="flex justify-center">
          <GoogleLogin
            onSuccess={handleSuccess}
            onError={handleError}
            useOneTap
          />
        </div>
      </div>
    </div>
  );
}

export default function Login() {
  return (
    <GoogleOAuthProvider clientId={import.meta.env.VITE_GOOGLE_CLIENT_ID || 'demo-client-id'}>
      <LoginContent />
    </GoogleOAuthProvider>
  );
}
