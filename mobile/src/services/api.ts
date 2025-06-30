import axios from 'axios';

// Configure the base URL for your Go backend
const API_BASE_URL = 'http://localhost:8080/api';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    console.log(`Making ${config.method?.toUpperCase()} request to ${config.url}`);
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response) {
      console.error('API Error:', error.response.data);
    } else if (error.request) {
      console.error('Network Error:', error.request);
    } else {
      console.error('Error:', error.message);
    }
    return Promise.reject(error);
  }
);

// API endpoints
export const roomGridAPI = {
  getRoomGrid: (propertyId: string, date: string) => 
    apiClient.get(`/room-grid/${propertyId}?date=${date}`),
  markBreakfastConsumed: (data: any) => 
    apiClient.post('/room-grid/consume', data),
  syncFromPMS: (propertyId: string) => 
    apiClient.post(`/room-grid/sync/${propertyId}`),
  getConsumptionHistory: (propertyId?: string, startDate?: string, endDate?: string) => {
    let url = '/room-grid/history';
    const params = new URLSearchParams();
    
    if (propertyId) params.append('property_id', propertyId);
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    
    const queryString = params.toString();
    return apiClient.get(queryString ? `${url}?${queryString}` : url);
  },
  getDailyReport: (propertyId: string, date: string) => 
    apiClient.get(`/room-grid/report/${propertyId}?date=${date}`),
};

export const authAPI = {
  login: (email: string, password: string) => 
    apiClient.post('/auth/login', { email, password }),
  register: (userData: any) => 
    apiClient.post('/auth/register', userData),
  logout: () => 
    apiClient.post('/auth/logout'),
  getProfile: () => 
    apiClient.get('/auth/me'),
};

export const analyticsAPI = {
  getDailyReport: (propertyId: string, date: string) => 
    roomGridAPI.getDailyReport(propertyId, date),
  getWeeklyReport: (propertyId: string, startDate: string) => 
    apiClient.get(`/room-grid/report/weekly/${propertyId}?start_date=${startDate}`),
  getMonthlyReport: (propertyId: string, month: string) => 
    apiClient.get(`/room-grid/report/monthly/${propertyId}?month=${month}`),
};
