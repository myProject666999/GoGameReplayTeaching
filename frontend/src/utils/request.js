import axios from 'axios'
import { useUserStore } from '@/stores/user'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

const PUBLIC_PATHS = ['/auth/login', '/auth/register']

request.interceptors.request.use(
  (config) => {
    const isPublic = PUBLIC_PATHS.some(p => config.url && config.url.includes(p))
    if (!isPublic) {
      const userStore = useUserStore()
      if (userStore.token) {
        config.headers.Authorization = `Bearer ${userStore.token}`
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response) {
      const url = error.config?.url || ''
      const isPublic = PUBLIC_PATHS.some(p => url.includes(p))
      if (error.response.status === 401 && !isPublic) {
        const userStore = useUserStore()
        userStore.logout()
      }
      const data = error.response.data
      if (data && data.error) {
        return Promise.reject(new Error(data.error))
      }
      return Promise.reject(new Error(`HTTP ${error.response.status}`))
    }
    return Promise.reject(error)
  }
)

export default request
