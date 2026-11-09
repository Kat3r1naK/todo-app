import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import LoginView from './views/LoginView.vue'
import BoardView from './views/BoardView.vue'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'board', component: BoardView, meta: { requiresAuth: true } },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const hasToken = !!localStorage.getItem('todo.token')
  if (to.meta.requiresAuth && !hasToken) {
    return { path: '/login' }
  }
  if (to.path === '/login' && hasToken && !to.query.key) {
    return { path: '/' }
  }
  return true
})

export default router
