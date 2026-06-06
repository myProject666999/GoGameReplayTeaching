import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/',
    name: 'GameList',
    component: () => import('@/views/GameList.vue'),
    meta: { title: '棋谱列表' }
  },
  {
    path: '/games/:id',
    name: 'GameDetail',
    component: () => import('@/views/GameDetail.vue'),
    meta: { title: '棋谱详情' }
  },
  {
    path: '/games/upload',
    name: 'GameUpload',
    component: () => import('@/views/GameUpload.vue'),
    meta: { title: '上传棋谱', requiresAuth: true }
  },
  {
    path: '/problems',
    name: 'ProblemList',
    component: () => import('@/views/ProblemList.vue'),
    meta: { title: '死活题列表' }
  },
  {
    path: '/problems/:id',
    name: 'ProblemDetail',
    component: () => import('@/views/ProblemDetail.vue'),
    meta: { title: '死活题详情' }
  },
  {
    path: '/problems/create',
    name: 'ProblemCreate',
    component: () => import('@/views/ProblemCreate.vue'),
    meta: { title: '死活题出题', requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { title: '注册' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router
