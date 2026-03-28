import { defineStore } from 'pinia'
import client from '../api/client'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    checked: false,
  }),
  actions: {
    setToken(t) {
      this.token = t
      if (t) localStorage.setItem('token', t)
      else localStorage.removeItem('token')
    },
    markChecked(v = true) {
      this.checked = v
    },
    async login(username, password) {
      const { data } = await client.post('/login', { username, password })
      this.setToken(data.token)
      this.markChecked(true)
      return data
    },
    async verifyToken() {
      if (!this.token) {
        this.markChecked(true)
        return false
      }
      try {
        await client.get('/home')
        this.markChecked(true)
        return true
      } catch (_) {
        this.logout()
        this.markChecked(true)
        return false
      }
    },
    logout() {
      this.setToken('')
      this.markChecked(true)
    },
  },
})
