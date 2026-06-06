<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <div class="logo">♟</div>
        <h1>围棋棋谱教学</h1>
        <p class="subtitle">登录您的账号</p>
      </div>
      <div class="form-group">
        <label>用户名</label>
        <input v-model="username" placeholder="请输入用户名" />
      </div>
      <div class="form-group">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="请输入密码" @keyup.enter="submit" />
      </div>
      <button @click="submit" :disabled="loading" class="btn-primary">
        {{ loading ? '登录中...' : '登录' }}
      </button>
      <transition name="fade">
        <p v-if="error" class="error">{{ error }}</p>
      </transition>
      <div class="divider">
        <span>或者</span>
      </div>
      <p class="link">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  if (!username.value.trim() || !password.value.trim()) {
    error.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await userStore.login(username.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 80px);
  padding: 40px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.auth-card {
  background: white;
  padding: 40px 36px;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
  width: 100%;
  max-width: 420px;
}
.auth-header {
  text-align: center;
  margin-bottom: 32px;
}
.logo {
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36px;
  color: white;
}
.auth-header h1 {
  font-size: 24px;
  color: #1a1a2e;
  margin-bottom: 6px;
}
.subtitle {
  color: #888;
  font-size: 14px;
}
.form-group {
  margin-bottom: 20px;
}
.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #374151;
  font-size: 14px;
}
.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 15px;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}
.form-group input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 4px rgba(102,126,234,0.1);
}
.btn-primary {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  margin-top: 8px;
  transition: transform 0.2s, box-shadow 0.2s;
}
.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102,126,234,0.4);
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.error {
  background: #fef2f2;
  color: #dc2626;
  padding: 10px 14px;
  border-radius: 8px;
  margin-top: 16px;
  font-size: 14px;
  text-align: center;
}
.divider {
  display: flex;
  align-items: center;
  margin: 24px 0 16px;
}
.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e5e7eb;
}
.divider span {
  padding: 0 16px;
  color: #9ca3af;
  font-size: 13px;
}
.link {
  text-align: center;
  color: #6b7280;
  font-size: 14px;
}
.link a {
  color: #667eea;
  text-decoration: none;
  font-weight: 600;
}
.link a:hover {
  text-decoration: underline;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
