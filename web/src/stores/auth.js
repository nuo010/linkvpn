import { defineStore } from 'pinia'
import client from '../api/client'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
  }),
  actions: {
    setToken(t) {
      this.token = t
      if (t) localStorage.setItem('token', t)
      else localStorage.removeItem('token')
    },
    async login(username, password) {
      const { data } = await client.post('/login', { username, password })
      this.setToken(data.token)
      return data
    },
    logout() {
      this.setToken('')
    },
  },
})
