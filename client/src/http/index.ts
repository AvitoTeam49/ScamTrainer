import axios from "axios";

export const API_URL = import.meta.env.VITE_API_URL ?? "/api/v1";

const $api = axios.create({
    baseURL: API_URL,
    withCredentials: true
});

$api.interceptors.request.use(config => {
    const token =
        localStorage.getItem("token");

    if (token) {
        config.headers.Authorization =
            `Bearer ${token}`;
    }

    return config;
});

$api.interceptors.response.use(
    response => response,
    async error => {
        const originalRequest = error.config;
        if (error.response?.status === 401 && originalRequest && !originalRequest._isRetry) {
            originalRequest._isRetry = true;
            try {
                const response = await axios.post(`${API_URL}/auth/refresh`, {}, {withCredentials: true});
                const newToken = response.data.access_token;
                localStorage.setItem("token", newToken);

                originalRequest.headers.Authorization = `Bearer ${newToken}`;

                return $api.request(originalRequest);

            } catch (refreshError) {
                localStorage.removeItem("token");

                window.location.href = "/auth";

                throw refreshError;
            }
        }

        throw error;
    }
);

export default $api;