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
        redirect: () => ({ path: '/settings/config' }),
        children: [
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
  strict: false, // 同时匹配 /settings 与 /settings/
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.public) return next()
  if (to.meta.requiresAuth && !auth.token) return next({ path: '/login', replace: true })
  // 统一去掉尾部斜杠，避免 /settings/ 与子路由重定向产生循环
  if (to.path.length > 1 && to.path.endsWith('/')) {
    return next({ path: to.path.slice(0, -1), replace: true })
  }
  next()
})

export default router
