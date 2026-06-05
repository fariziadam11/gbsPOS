import axios from 'axios'

const posApiClient = axios.create({
  baseURL: import.meta.env.VITE_POS_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

posApiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

posApiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const isLoginRequest = error.config?.url?.endsWith('/login')
    if (error.response?.status === 401 && !isLoginRequest) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default posApiClient
