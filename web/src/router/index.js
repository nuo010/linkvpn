import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', name: 'Login', component: () => import('../views/Login.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/home' },
      { path: 'home', name: 'home', component: () => import('../views/Home.vue') },
      { path: 'user', name: 'user', component: () => import('../views/UserList.vue') },
      {
        path: 'settings',
        component: () => import('../views/Settings.vue'),
        children: [
          { path: '', redirect: '/settings/config' },
          { path: 'config', name: 'settingsConfig', component: () => import('../views/Config.vue') },
          { path: 'system', name: 'settingsSystem', component: () => import('../views/System.vue') },
        ],
      },
      { path: 'config', redirect: '/settings/config' },
      { path: 'system', redirect: '/settings/system' },
      { path: 'logs/login', name: 'logsLogin', component: () => import('../views/LogsLogin.vue') },
      { path: 'logs/vpn', redirect: '/logs/main' },
      { path: 'logs/main', name: 'logMain', component: () => import('../views/LogView.vue'), meta: { logFile: 'openvpn.log', title: '服务日志' } },
      { path: 'logs/status', name: 'logStatus', component: () => import('../views/LogView.vue'), meta: { logFile: 'openvpn-status.log', title: '连接日志' } },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (to.path === '/login' && auth.token && !auth.checked) {
      const ok = await auth.verifyToken()
      if (ok) return next({ path: '/', replace: true })
    }
    return next()
  }
  if (!auth.token) return next({ path: '/login', replace: true })
  if (!auth.checked) {
    const ok = await auth.verifyToken()
    if (!ok) return next({ path: '/login', replace: true })
  }
  next()
})

export default router
