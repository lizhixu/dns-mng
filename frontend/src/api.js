// API Base URL - 根据环境自动选择
const getApiBase = () => {
    // 生产环境：使用相对路径（通过 Nginx 代理）
    if (import.meta.env.PROD) {
        return '/api';
    }
    // 开发环境：直接连接后端
    return import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
};

const API_BASE = getApiBase();

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
};
