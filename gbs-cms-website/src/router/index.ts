import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AdminLayout from '../layouts/AdminLayout.vue'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import ResourceView from '../views/ResourceView.vue'
import QrisView from '../views/QrisView.vue'
import SettingsView from '../views/SettingsView.vue'

const routes = [
  { path: '/login', component: LoginView, meta: { guestOnly: true } },
  {
    path: '/', component: AdminLayout, meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: DashboardView },
      { path: 'products', component: ResourceView, meta: { title: 'Products', resource: 'products' } },
      { path: 'orders', component: ResourceView, meta: { title: 'Orders', resource: 'orders' } },
      { path: 'settlements', component: ResourceView, meta: { title: 'Settlements', resource: 'settlements' } },
      { path: 'customers', component: ResourceView, meta: { title: 'Customers', resource: 'customers' } },
      { path: 'discounts', component: ResourceView, meta: { title: 'Discounts', resource: 'discounts' } },
      { path: 'ads', component: ResourceView, meta: { title: 'Advertisements', resource: 'ads' } },
      { path: 'users', component: ResourceView, meta: { title: 'Users', resource: 'users' } },
      { path: 'settings', component: SettingsView, meta: { title: 'Settings', resource: 'settings' } },
      { path: 'display', component: ResourceView, meta: { title: 'Display Terminals', resource: 'display' } },
      { path: 'qris', component: QrisView, meta: { title: 'QRIS Payments', resource: 'qris' } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({ history: createWebHistory(), routes })
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) return { path: '/login', query: { redirect: to.fullPath } }
  if (to.meta.guestOnly && auth.isAuthenticated) return '/dashboard'
})
window.addEventListener('gbs:unauthorized', () => { useAuthStore().logout(); void router.push('/login') })
export default router
