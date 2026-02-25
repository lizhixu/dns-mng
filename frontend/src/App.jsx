import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './AuthContext';
import { LanguageProvider } from './LanguageContext';
import { ThemeProvider } from './ThemeContext';
import Layout from './components/Layout';
import Login from './pages/Login';
import Accounts from './pages/Accounts';
import AllDomains from './pages/AllDomains';
import Domains from './pages/Domains';
import Records from './pages/Records';
import Profile from './pages/Profile';
import Logs from './pages/Logs';

// Placeholder components until we implement them
const PrivateRoute = ({ children }) => {
  const token = localStorage.getItem('token');
  return token ? children : <Navigate to="/login" replace />;
};

function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <LanguageProvider>
          <AuthProvider>
            <Routes>
              <Route path="/login" element={<Login />} />

              <Route path="/" element={
                <PrivateRoute>
                  <Layout />
                </PrivateRoute>
              }>
                <Route index element={<Navigate to="/domains" replace />} />
                <Route path="domains" element={<AllDomains />} />
                <Route path="accounts" element={<Accounts />} />
                <Route path="accounts/:accountId/domains" element={<Domains />} />
                <Route path="accounts/:accountId/domains/:domainId/records" element={<Records />} />
                <Route path="profile" element={<Profile />} />
                <Route path="logs" element={<Logs />} />
              </Route>
            </Routes>
          </AuthProvider>
        </LanguageProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}

export default App;
