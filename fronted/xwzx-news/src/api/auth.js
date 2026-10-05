import axios from 'axios';
import { useUserStore } from '../store/user';
import { apiConfig } from '../config/api';

let refreshPromise;
const authPaths = ['/api/user/login', '/api/user/register', '/api/user/refresh', '/api/user/logout'];
const refreshLockName = 'news-py-go-refresh-token';

function withRefreshLock(operation) {
  const lockManager = globalThis.navigator?.locks;
  if (!lockManager) return operation();
  return lockManager.request(refreshLockName, { mode: 'exclusive' }, operation);
}

function isAuthRequest(config) {
  return authPaths.some((path) => config.url?.endsWith(path));
}

export async function refreshAccessToken(userStore) {
  if (!refreshPromise) {
    refreshPromise = withRefreshLock(() =>
      axios.post(`${apiConfig.baseURL}/api/user/refresh`, {}, { withCredentials: true })
      .then((response) => {
        if (response.data?.code !== 200 || !response.data.data?.token) {
          throw new Error(response.data?.message || '登录状态已失效');
        }
        userStore.setSession(response.data.data);
        return response.data.data.token;
      })
      .catch((error) => {
        if (error.response?.status === 401) userStore.clearSession();
        throw error;
      })
    ).finally(() => {
      refreshPromise = undefined;
    });
  }
  return refreshPromise;
}

export function installAuthInterceptors(pinia) {
  const userStore = useUserStore(pinia);

  axios.interceptors.request.use((config) => {
    if (isAuthRequest(config)) config.withCredentials = true;
    if (userStore.token && !isAuthRequest(config)) {
      config.headers = config.headers || {};
      config.headers.Authorization = userStore.token;
    }
    return config;
  });

  axios.interceptors.response.use(undefined, async (error) => {
    const config = error.config;
    if (error.response?.status !== 401 || !config || config._authRetried || isAuthRequest(config)) {
      return Promise.reject(error);
    }
    config._authRetried = true;
    try {
      const token = await refreshAccessToken(userStore);
      config.headers = config.headers || {};
      config.headers.Authorization = token;
      return axios(config);
    } catch {
      return Promise.reject(error);
    }
  });
}

export async function restoreSession(userStore) {
  if (userStore.suppressRefreshRestore) return;
  try {
    await refreshAccessToken(userStore);
  } catch {}
}

export async function fetchWithAuth(input, init = {}) {
  const userStore = useUserStore();
  const send = () => fetch(input, {
    ...init,
    headers: { ...init.headers, ...(userStore.token ? { Authorization: userStore.token } : {}) }
  });
  let response = await send();
  if (response.status !== 401) return response;
  try {
    const token = await refreshAccessToken(userStore);
    response = await fetch(input, {
      ...init,
      headers: { ...init.headers, Authorization: token }
    });
  } catch {
    userStore.clearSession();
  }
  return response;
}
