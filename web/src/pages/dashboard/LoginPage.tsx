import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../../api/client';

export function LoginPage() {
  const [apiKey, setApiKey] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async () => {
    sessionStorage.setItem('api_key', apiKey);
    try {
      await api.getProfile();
      navigate('/dashboard');
    } catch {
      setError('Invalid API key');
      sessionStorage.removeItem('api_key');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-lg p-8 max-w-sm w-full">
        <h1 className="text-2xl font-bold text-gray-900 mb-6 text-center">Merchant Login</h1>
        <input
          type="password"
          placeholder="Enter your API Key"
          value={apiKey}
          onChange={e => { setApiKey(e.target.value); setError(''); }}
          onKeyDown={e => e.key === 'Enter' && handleLogin()}
          className="w-full border rounded-lg px-4 py-3 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
        />
        {error && <p className="text-red-500 text-sm mt-2">{error}</p>}
        <button
          onClick={handleLogin}
          className="w-full mt-4 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition"
        >
          Login
        </button>
      </div>
    </div>
  );
}
