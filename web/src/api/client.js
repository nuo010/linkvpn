import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

client.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      // 使用 replace 避免历史记录里留下会再次触发 401 的地址，减少重定向循环风险
      window.location.replace('/login')
    }
    return Promise.reject(err)
  }
)

export default client
