<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const menuOpen = ref(false)
const active = computed(() => route.meta.resource as string | undefined)
const pageTitle = computed(() => (route.meta.title as string | undefined) ?? 'Overview')

const groups = [
  { label: 'Workspace', links: [{ label: 'Dashboard', to: '/dashboard', resource: undefined, icon: '⌂' }] },
  { label: 'POS operations', links: [
    { label: 'Products', to: '/products', resource: 'products', icon: '▤' },
    { label: 'Orders', to: '/orders', resource: 'orders', icon: '↗' },
    { label: 'Settlements', to: '/settlements', resource: 'settlements', icon: '◫' },
    { label: 'Customers', to: '/customers', resource: 'customers', icon: '♙' },
    { label: 'Discounts', to: '/discounts', resource: 'discounts', icon: '%' },
  ] },
  { label: 'CMS & devices', links: [
    { label: 'Advertisements', to: '/ads', resource: 'ads', icon: '▷' },
    { label: 'Display terminals', to: '/display', resource: 'display', icon: '▣' },
    { label: 'QRIS payments', to: '/qris', resource: 'qris', icon: '⌁' },
  ] },
  { label: 'Administration', links: [{ label: 'Users', to: '/users', resource: 'users', icon: '♟' }, { label: 'Settings', to: '/settings', resource: 'settings', icon: '⚙' }] },
]
</script>

<template>
  <div class="min-h-screen bg-[#f4f7f8] text-slate-900 lg:flex">
    <div v-if="menuOpen" class="fixed inset-0 z-30 bg-slate-950/50 lg:hidden" @click="menuOpen = false" />
    <aside :class="['fixed inset-y-0 left-0 z-40 flex w-[280px] -translate-x-full flex-col bg-[#102a2b] text-slate-300 shadow-2xl transition-transform lg:translate-x-0', menuOpen ? 'translate-x-0' : '']">
      <div class="flex items-start justify-between px-6 pb-7 pt-7"><div><div class="flex items-center gap-2"><span class="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 font-black text-white">G</span><span class="text-lg font-bold tracking-tight text-white">GBS</span></div><p class="mt-3 text-[11px] font-semibold uppercase tracking-[.24em] text-teal-300/70">Operations suite</p></div><button class="text-xl text-slate-400 lg:hidden" aria-label="Close menu" @click="menuOpen = false">×</button></div>
      <nav class="min-h-0 flex-1 space-y-6 overflow-y-auto px-4 pb-6"><div v-for="group in groups" :key="group.label"><p class="mb-2 px-3 text-[10px] font-bold uppercase tracking-[.18em] text-slate-500">{{ group.label }}</p><div class="space-y-1"><RouterLink v-for="link in group.links" :key="link.to" :to="link.to" :class="['nav-link', (active === link.resource || (!active && !link.resource)) ? 'nav-link-active' : '']" @click="menuOpen = false"><span class="nav-icon">{{ link.icon }}</span><span>{{ link.label }}</span></RouterLink></div></div></nav>
      <div class="m-4 rounded-2xl border border-white/10 bg-white/5 p-4"><div class="flex items-center gap-3"><span class="grid h-9 w-9 place-items-center rounded-full bg-teal-300 font-bold text-[#102a2b]">{{ auth.user?.name?.charAt(0).toUpperCase() }}</span><div class="min-w-0"><p class="truncate text-sm font-semibold text-white">{{ auth.user?.name }}</p><p class="mt-0.5 text-xs text-slate-400">{{ auth.user?.role }}</p></div></div></div>
    </aside>
    <main class="min-w-0 w-full lg:ml-[280px]">
      <header class="sticky top-0 z-20 border-b border-slate-200/80 bg-[#f4f7f8]/90 px-5 py-4 backdrop-blur-xl sm:px-8"><div class="flex items-center justify-between gap-4"><div class="flex items-center gap-3"><button class="grid h-10 w-10 place-items-center rounded-xl border border-slate-200 bg-white text-xl text-slate-700 lg:hidden" aria-label="Open menu" @click="menuOpen = true">☰</button><div><div class="flex items-center gap-2 text-xs font-semibold text-slate-400"><span>GBS</span><span>/</span><span class="text-brand-600">{{ pageTitle }}</span></div><h1 class="mt-1 text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">{{ pageTitle }}</h1></div></div><div class="flex items-center gap-3"><div class="hidden text-right sm:block"><p class="text-xs font-semibold text-slate-700">{{ auth.user?.name }}</p><p class="text-[11px] text-slate-400">{{ auth.user?.role }} access</p></div><button class="grid h-10 w-10 place-items-center rounded-xl border border-slate-200 bg-white text-sm font-bold text-brand-700 shadow-sm hover:border-brand-200" title="Sign out" aria-label="Sign out" @click="auth.logout">{{ auth.user?.name?.charAt(0).toUpperCase() }}</button></div></div></header>
      <div class="mx-auto max-w-[1500px] p-5 sm:p-8"><RouterView /></div>
    </main>
  </div>
</template>
