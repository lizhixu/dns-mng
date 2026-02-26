// API Base URL - 直接访问后端
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

const getHeaders = () => {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
};

const handleResponse = async (response) => {
    const contentType = response.headers.get('content-type');
    const isJson = contentType && contentType.includes('application/json');
    const data = isJson ? await response.json() : await response.text();

    if (!response.ok) {
        if (response.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login';
        }
        const error = (data && data.error) || response.statusText;
        throw new Error(error);
    }
    return data;
};

export const api = {
    // Auth
    login: async (username, password) => {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
        });
        return handleResponse(response);
    },

    register: async (username, password) => {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
        });
        return handleResponse(response);
    },

    // User Profile
    getProfile: async () => {
        const response = await fetch(`${API_BASE}/user/profile`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    updatePassword: async (data) => {
        const response = await fetch(`${API_BASE}/user/password`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    // Operation Logs
    getLogs: async (limit = 50) => {
        const response = await fetch(`${API_BASE}/logs?limit=${limit}`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Providers
    getProviders: async () => {
        const response = await fetch(`${API_BASE}/providers`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Accounts
    getAccounts: async () => {
        const response = await fetch(`${API_BASE}/accounts`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    createAccount: async (data) => {
        const response = await fetch(`${API_BASE}/accounts`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    updateAccount: async (id, data) => {
        const response = await fetch(`${API_BASE}/accounts/${id}`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    deleteAccount: async (id) => {
        const response = await fetch(`${API_BASE}/accounts/${id}`, {
            method: 'DELETE',
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // All Domains
    getAllDomains: async () => {
        const response = await fetch(`${API_BASE}/domains`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Domains
    getDomains: async (accountId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    getDomain: async (accountId, domainId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Records
    getRecords: async (accountId, domainId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/records`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    createRecord: async (accountId, domainId, data) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/records`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    updateRecord: async (accountId, domainId, recordId, data) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/records/${recordId}`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    deleteRecord: async (accountId, domainId, recordId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/records/${recordId}`, {
            method: 'DELETE',
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // DNS Check
    checkDNS: async (data) => {
        const response = await fetch(`${API_BASE}/dns/check`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },
};
