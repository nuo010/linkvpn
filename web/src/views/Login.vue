<template>
  <div class="login-page">
    <div class="login-bg" aria-hidden="true">
      <div class="login-bg-sky"></div>
      <div class="login-bg-clouds"></div>
      <div class="login-bg-sun"></div>
    </div>
    <div class="login-card">
      <div class="login-brand">
        <img class="login-logo" :src="logoUrl" alt="LinkVPN" />
        <span class="login-name">LinkVPN</span>
      </div>
      <p class="login-desc">登录以管理 VPN 用户与配置</p>
      <form @submit.prevent="submit" class="login-form">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="username" type="text" required autocomplete="username" placeholder="请输入用户名" />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="password" type="password" required autocomplete="current-password" placeholder="请输入密码" />
        </div>
        <div v-if="error" class="login-error">{{ error }}</div>
        <button type="submit" class="login-btn" :disabled="loading">{{ loading ? '登录中…' : '登录' }}</button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import logoUrl from '../logo.png'

const router = useRouter()
const auth = useAuthStore()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  position: relative;
  overflow: hidden;
}

/* 背景：晴空偏白蓝 — 上浅蓝下近白 + 柔和云团 + 顶部日光感 */
.login-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  background: linear-gradient(180deg, #c8e7ff 0%, #e8f4fc 35%, #f5faff 65%, #fbfdff 100%);
}

.login-bg-sky {
  position: absolute;
  inset: 0;
  /* 更通透的天蓝层次 */
  background:
    linear-gradient(180deg, rgba(125, 211, 252, 0.45) 0%, transparent 42%),
    linear-gradient(165deg, transparent 50%, rgba(224, 242, 254, 0.6) 100%);
  pointer-events: none;
}

.login-bg-clouds {
  position: absolute;
  inset: 0;
  /* 云团：大块柔边白 + 淡蓝，模拟蓝天下的体积感 */
  background:
    radial-gradient(ellipse 90% 55% at 15% 25%, rgba(255, 255, 255, 0.92) 0%, rgba(255, 255, 255, 0) 60%),
    radial-gradient(ellipse 70% 45% at 85% 18%, rgba(255, 255, 255, 0.75) 0%, rgba(255, 255, 255, 0) 55%),
    radial-gradient(ellipse 100% 50% at 50% 8%, rgba(255, 255, 255, 0.55) 0%, rgba(255, 255, 255, 0) 50%),
    radial-gradient(ellipse 80% 40% at 70% 45%, rgba(186, 230, 253, 0.5) 0%, transparent 55%),
    radial-gradient(ellipse 60% 35% at 20% 55%, rgba(224, 242, 254, 0.65) 0%, transparent 50%),
    radial-gradient(ellipse 110% 60% at 50% 100%, rgba(255, 255, 255, 0.85) 0%, rgba(255, 255, 255, 0) 45%);
  pointer-events: none;
  animation: login-cloud-drift 24s ease-in-out infinite alternate;
}

@keyframes login-cloud-drift {
  0% {
    transform: translate(0, 0);
    opacity: 1;
  }
  100% {
    transform: translate(2%, -1%);
    opacity: 0.97;
  }
}

.login-bg-sun {
  position: absolute;
  inset: 0;
  /* 顶部柔和「日光」 */
  background: radial-gradient(ellipse 70% 45% at 50% -5%, rgba(255, 255, 255, 0.9) 0%, rgba(255, 255, 255, 0.15) 35%, transparent 60%);
  pointer-events: none;
}

.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
  padding: 2.5rem 2.25rem;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 18px;
  border: 1px solid rgba(186, 230, 253, 0.8);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.8) inset,
    0 20px 50px -12px rgba(14, 165, 233, 0.12),
    0 12px 32px -8px rgba(56, 189, 248, 0.08);
  text-align: center;
  backdrop-filter: blur(12px);
}
.login-brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.login-logo {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  object-fit: contain;
  background: #fff;
  border: 1px solid rgba(186, 230, 253, 0.9);
  padding: 6px;
  box-shadow: 0 8px 24px rgba(14, 165, 233, 0.1);
}
.login-name {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.02em;
}
.login-desc {
  color: #64748b;
  font-size: 0.9rem;
  margin: 0 0 1.75rem;
}
.login-form .form-group {
  margin-bottom: 1.1rem;
  text-align: left;
}
.login-form .form-group label {
  display: block;
  margin-bottom: 0.35rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #334155;
}
.login-form .form-group input {
  width: 100%;
  height: 44px;
  padding: 0 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.95rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}
.login-form .form-group input:focus {
  outline: none;
  border-color: #0ea5e9;
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.15);
}
.login-form .form-group input::placeholder {
  color: #94a3b8;
}
.login-error {
  margin-top: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: #fef2f2;
  color: #dc2626;
  font-size: 0.875rem;
  border-radius: 8px;
  text-align: left;
}
.login-btn {
  width: 100%;
  margin-top: 1rem;
  height: 44px;
  padding: 0 1.5rem;
  background: linear-gradient(135deg, #0ea5e9 0%, #06b6d4 100%);
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.15s;
}
.login-btn:hover:not(:disabled) {
  opacity: 0.95;
  transform: translateY(-1px);
}
.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}
</style>
