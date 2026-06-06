<template>
  <div class="app-container">
    <nav class="navbar">
      <div class="nav-brand">
        <router-link to="/">围棋教学系统</router-link>
      </div>
      <div class="nav-links">
        <router-link to="/">棋谱列表</router-link>
        <router-link to="/games/upload">上传棋谱</router-link>
        <router-link to="/problems">死活题列表</router-link>
        <router-link to="/problems/create">出题</router-link>
        <template v-if="!userStore.isLoggedIn">
          <router-link to="/login">登录</router-link>
          <router-link to="/register">注册</router-link>
        </template>
        <template v-else>
          <span class="user-info">{{ userStore.user?.nickname || userStore.user?.username }}</span>
          <button @click="userStore.logout()">退出</button>
        </template>
      </div>
    </nav>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: #f5f5f5;
}

.app-container {
  min-height: 100vh;
}

.navbar {
  background-color: #2c3e50;
  color: white;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
}

.nav-brand a {
  color: white;
  text-decoration: none;
  font-size: 20px;
  font-weight: bold;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 20px;
}

.nav-links a {
  color: white;
  text-decoration: none;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  background-color: #34495e;
}

.user-info {
  color: #ecf0f1;
}

.nav-links button {
  background-color: #e74c3c;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
}

.nav-links button:hover {
  background-color: #c0392b;
}

.main-content {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}
</style>
