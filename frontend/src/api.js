// API Base URL - 直接访问后端
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Get backend base URL (without /api suffix for DDNS)
export const getBackendBaseURL = () => {
    const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
    // Remove /api suffix if present
    return apiBase.replace(/\/api$/, '');
};

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
    getLogs: async (page = 1, pageSize = 20) => {
        const response = await fetch(`${API_BASE}/logs?page=${page}&page_size=${pageSize}`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Scheduler Logs
    getSchedulerLogs: async (page = 1, pageSize = 20) => {
        const response = await fetch(`${API_BASE}/scheduler-logs?page=${page}&page_size=${pageSize}`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    getSchedulerLogsByTask: async (taskName, page = 1, pageSize = 20) => {
        const response = await fetch(`${API_BASE}/scheduler-logs/${taskName}?page=${page}&page_size=${pageSize}`, {
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

    // Refresh all domains from providers
    refreshAllDomains: async () => {
        const response = await fetch(`${API_BASE}/domains/refresh`, {
            headers: getHeaders(),
        });
        const data = await handleResponse(response);
        return data;
    },

    // Domains
    getDomains: async (accountId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Refresh domains from provider
    refreshDomains: async (accountId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/refresh`, {
            headers: getHeaders(),
        });
        const data = await handleResponse(response);
        // Return the full response object with domains, domains_to_delete, and cache_timestamp
        return data;
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

    // Domain Cache (renewal info)
    updateDomainCache: async (accountId, domainId, data) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/cache`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    // Batch update domain cache
    batchUpdateDomainCache: async (items) => {
        const response = await fetch(`${API_BASE}/cache/batch`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ items }),
        });
        return handleResponse(response);
    },

    // Batch delete domain cache
    batchDeleteDomainCache: async (items) => {
        const response = await fetch(`${API_BASE}/cache/batch`, {
            method: 'DELETE',
            headers: getHeaders(),
            body: JSON.stringify({ items }),
        });
        return handleResponse(response);
    },

    // Get cache statistics
    getCacheStats: async () => {
        const response = await fetch(`${API_BASE}/cache/stats`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Notification settings
    getNotificationSetting: async (accountId, domainId) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/notification`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    updateNotificationSetting: async (accountId, domainId, data) => {
        const response = await fetch(`${API_BASE}/accounts/${accountId}/domains/${domainId}/notification`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    getAllNotificationSettings: async () => {
        const response = await fetch(`${API_BASE}/notifications`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Email configuration
    getEmailConfig: async () => {
        const response = await fetch(`${API_BASE}/email/config`, {
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    updateEmailConfig: async (data) => {
        const response = await fetch(`${API_BASE}/email/config`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data),
        });
        return handleResponse(response);
    },

    testEmailConfig: async () => {
        const response = await fetch(`${API_BASE}/email/test`, {
            method: 'POST',
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Trigger manual scheduler check
    triggerSchedulerCheck: async () => {
        const response = await fetch(`${API_BASE}/scheduler/trigger`, {
            method: 'POST',
            headers: getHeaders(),
        });
        return handleResponse(response);
    },

    // Batch soft delete domains
    batchSoftDeleteDomains: async (items) => {
        const response = await fetch(`${API_BASE}/domains/batch-soft-delete`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ items }),
        });
        return handleResponse(response);
    },

    // Batch restore domains
    batchRestoreDomains: async (items) => {
        const response = await fetch(`${API_BASE}/domains/batch-restore`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ items }),
        });
        return handleResponse(response);
    },

    // DDNS Token API (user-level, one token per user)
    getDDNSToken: async () => {
        const response = await fetch(`${API_BASE}/ddns-token`, {
            headers: getHeaders()
        });
        return handleResponse(response);
    },

    updateDDNSToken: async (data) => {
        const response = await fetch(`${API_BASE}/ddns-token`, {
            method: 'PUT',
            headers: getHeaders(),
            body: JSON.stringify(data)
        });
        return handleResponse(response);
    },

    deleteDDNSToken: async () => {
        const response = await fetch(`${API_BASE}/ddns-token`, {
            method: 'DELETE',
            headers: getHeaders()
        });
        return handleResponse(response);
    },
};
