import { BrowserRouter, Routes, Route, Link, Navigate, Outlet } from 'react-router-dom';
import { CheckoutPage } from './pages/checkout/CheckoutPage';
import { LoginPage } from './pages/dashboard/LoginPage';
import { DashboardPage } from './pages/dashboard/DashboardPage';
import { PaymentsPage } from './pages/dashboard/PaymentsPage';
import { PayoutsPage } from './pages/dashboard/PayoutsPage';
import { SettingsPage } from './pages/dashboard/SettingsPage';
import { RefundsPage } from './pages/dashboard/RefundsPage';
import { SweepPage } from './pages/dashboard/SweepPage';

function DashboardLayout() {
  const isLoggedIn = !!sessionStorage.getItem('api_key');
  if (!isLoggedIn) return <Navigate to="/login" />;

  return (
    <div className="min-h-screen bg-gray-100">
      <nav className="fixed left-0 top-0 h-full w-56 bg-gray-900 text-white p-4 space-y-1">
        <h1 className="text-lg font-bold mb-6 px-3">OctopusWallet</h1>
        <NavLink to="/dashboard" label="Dashboard" />
        <NavLink to="/dashboard/payments" label="Payments" />
        <NavLink to="/dashboard/payouts" label="Payouts" />
        <NavLink to="/dashboard/refunds" label="Refunds" />
        <NavLink to="/dashboard/sweeps" label="Auto-Sweep" />
        <NavLink to="/dashboard/settings" label="Settings" />
        <div className="absolute bottom-4 left-4 right-4">
          <button
            onClick={() => { sessionStorage.removeItem('api_key'); window.location.href = '/login'; }}
            className="w-full text-left px-3 py-2 text-sm text-gray-400 hover:text-white"
          >
            Logout
          </button>
        </div>
      </nav>
      <main className="ml-56 p-8">
        <Outlet />
      </main>
    </div>
  );
}

function NavLink({ to, label }: { to: string; label: string }) {
  return (
    <Link to={to} className="block px-3 py-2 rounded text-sm text-gray-300 hover:bg-gray-800 hover:text-white transition">
      {label}
    </Link>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/pay/:id" element={<CheckoutPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="payments" element={<PaymentsPage />} />
          <Route path="payouts" element={<PayoutsPage />} />
          <Route path="refunds" element={<RefundsPage />} />
          <Route path="sweeps" element={<SweepPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>
        <Route path="/" element={<Navigate to="/login" />} />
      </Routes>
    </BrowserRouter>
  );
}
