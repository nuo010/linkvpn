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
      // 避免在登录页或重复跳转过程中再次 replace，减少循环跳转风险
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        window.location.replace('/login')
      }
    }
    return Promise.reject(err)
  }
)

export default client
