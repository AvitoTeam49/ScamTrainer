import axios from "axios";

export const API_URL = import.meta.env.VITE_API_URL ?? "/api/v1";

const AUTH_PATHS = ["/auth/login", "/auth/register", "/auth/refresh"];

const $api = axios.create({
    baseURL: API_URL,
    withCredentials: true
});

let refreshRequest: Promise<string> | null = null;

export const refreshAccessToken = (): Promise<string> => {
    if (!refreshRequest) {
        refreshRequest = axios
            .post<{ access_token: string }>(`${API_URL}/auth/refresh`, {}, {withCredentials: true})
            .then(response => {
                const token = response.data.access_token;

                localStorage.setItem("token", token);

                return token;
            })
            .finally(() => {
                refreshRequest = null;
            });
    }

    return refreshRequest;
};

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

        if (error.response?.status !== 401 || !originalRequest || originalRequest._isRetry) {
            throw error;
        }

        const url = originalRequest.url ?? "";

        if (AUTH_PATHS.some(path => url.startsWith(path))) {
            throw error;
        }

        originalRequest._isRetry = true;

        try {
            const token = await refreshAccessToken();

            originalRequest.headers.Authorization = `Bearer ${token}`;

            return await $api.request(originalRequest);

        } catch (refreshError) {
            localStorage.removeItem("token");

            window.location.href = "/auth";

            throw refreshError;
        }
    }
);

export default $api;
