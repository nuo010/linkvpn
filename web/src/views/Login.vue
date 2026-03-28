<template>
  <div class="login-page">
    <div class="login-bg" aria-hidden="true">
      <div class="login-bg-sky"></div>
      <div class="login-bg-clouds"></div>
      <div class="login-bg-sun"></div>
      <div class="login-bg-grid"></div>
    </div>
    <div class="login-shell">
      <div class="login-panel login-card">
        <div class="login-brand">
          <img class="login-logo" :src="logoUrl" alt="LinkVPN" />
          <div class="login-brand-copy">
            <span class="login-name">LinkVPN</span>
            <span class="login-subtitle">管理员登录</span>
          </div>
        </div>
        <p class="login-desc">登录以管理 VPN 用户、配置和日志</p>
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
  padding: 1.8rem;
  position: relative;
  overflow: hidden;
}

/* 背景：晴空偏白蓝 — 上浅蓝下近白 + 柔和云团 + 顶部日光感 */
.login-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  background: linear-gradient(180deg, #d9eeff 0%, #edf6ff 34%, #f7fbff 68%, #fcfdff 100%);
}

.login-bg-sky {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(125, 211, 252, 0.32) 0%, transparent 42%),
    linear-gradient(165deg, transparent 50%, rgba(224, 242, 254, 0.48) 100%);
  pointer-events: none;
}

.login-bg-clouds {
  position: absolute;
  inset: 0;
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
.login-bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.06) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.28), transparent 78%);
  pointer-events: none;
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
  background: radial-gradient(ellipse 70% 45% at 50% -5%, rgba(255, 255, 255, 0.9) 0%, rgba(255, 255, 255, 0.15) 35%, transparent 60%);
  pointer-events: none;
}
.login-shell {
  position: relative;
  z-index: 1;
  width: min(100%, 460px);
}
.login-panel {
  border-radius: 24px;
  border: 1px solid rgba(219, 231, 243, 0.92);
  backdrop-filter: blur(12px);
}
.login-card {
  width: 100%;
  padding: 2rem 1.9rem 1.85rem;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 24px;
  border: 1px solid rgba(219, 231, 243, 0.92);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.85) inset,
    0 24px 52px rgba(15, 23, 42, 0.1);
  backdrop-filter: blur(12px);
}
.login-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.65rem;
}
.login-brand-copy {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
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
  font-size: 1.6rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.02em;
}
.login-subtitle {
  color: #64748b;
  font-size: 0.86rem;
  font-weight: 600;
}
.login-desc {
  color: #64748b;
  font-size: 0.9rem;
  margin: 0 0 1.5rem;
}
.login-form .form-group {
  margin-bottom: 1.1rem;
  text-align: left;
}
.login-form .form-group label {
  display: block;
  margin-bottom: 0.38rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}
.login-form .form-group input {
  width: 100%;
  height: 46px;
  padding: 0 1rem;
  border: 1px solid #dbe5f1;
  border-radius: 12px;
  font-size: 0.95rem;
  background: rgba(255, 255, 255, 0.95);
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
  box-sizing: border-box;
}
.login-form .form-group input:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.14);
  background: #fff;
}
.login-form .form-group input::placeholder {
  color: #94a3b8;
}
.login-error {
  margin-top: 0.5rem;
  padding: 0.65rem 0.8rem;
  background: linear-gradient(180deg, #fff5f5 0%, #fef2f2 100%);
  color: #dc2626;
  font-size: 0.875rem;
  border-radius: 10px;
  text-align: left;
  border: 1px solid #fecaca;
}
.login-btn {
  width: 100%;
  margin-top: 1rem;
  height: 46px;
  padding: 0 1.5rem;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: #fff;
  border: 1px solid #2563eb;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.2);
  transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s;
}
.login-btn:hover:not(:disabled) {
  opacity: 0.97;
  transform: translateY(-1px);
  box-shadow: 0 18px 30px rgba(37, 99, 235, 0.24);
}
.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}
@media (max-width: 520px) {
  .login-page {
    padding: 1rem;
  }
  .login-card {
    padding: 1.35rem 1.15rem;
    border-radius: 18px;
  }
}
</style>
