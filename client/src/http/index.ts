import axios from 'axios';
import type {LoginResponse} from "../types/types.tsx";

export const API_URL = 'http://localhost:8080/api/v1';


const $api = axios.create({
    withCredentials: true,
    baseURL: API_URL
})

$api.interceptors.request.use(config => {
    config.headers.Authorization = `Bearer ${localStorage.getItem('token')}`;
    return config;
})


$api.interceptors.response.use((config) => {
    return config;
}, async (error) => {
    const originalRequest = error.config;
    if (error.response?.status == 401 && error.config && !error.config._isRetry) {
        originalRequest._isRetry = true
        try {
            const response = await axios.post<LoginResponse>(`${API_URL}/auth/refresh`,{},{withCredentials: true});
            localStorage.setItem('token', response.data.access_token);
            originalRequest.headers.Authorization = `Bearer ${response.data.access_token}`;
            return $api.request(originalRequest);
        } catch (e) {
            localStorage.removeItem("token");
            window.location.href = "/auth";
            throw e;
        }
    }
    throw error;
})

export default $api;


