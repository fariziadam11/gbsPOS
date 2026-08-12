<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const username = ref('admin')
const password = ref('admin123')
const error = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true; error.value = ''
  try { await auth.login(username.value, password.value); await router.push(String(route.query.redirect || '/dashboard')) }
  catch (err) { error.value = err instanceof Error ? err.message : 'Login gagal' }
  finally { loading.value = false }
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-slate-950 px-5 py-10"><div class="grid w-full max-w-4xl overflow-hidden rounded-3xl bg-white shadow-2xl md:grid-cols-2"><section class="hidden bg-brand-700 p-10 text-white md:block"><p class="text-xs font-semibold uppercase tracking-[.35em] text-brand-100">GBS POS + CMS</p><h1 class="mt-20 text-4xl font-bold leading-tight">Satu ruang kendali untuk seluruh operasional.</h1><p class="mt-5 text-brand-100">Pantau penjualan, katalog, konten, dan pengguna dari satu dashboard.</p></section><form class="p-7 sm:p-10" @submit.prevent="submit"><p class="text-sm font-semibold text-brand-600">Welcome back</p><h2 class="mt-2 text-3xl font-bold text-slate-900">Sign in</h2><p class="mt-2 text-sm text-slate-500">Masuk dengan akun administrator Anda.</p><div v-if="error" class="mt-6 rounded-xl bg-red-50 p-3 text-sm text-red-700">{{ error }}</div><label class="mt-7 block text-sm font-medium text-slate-700">Username<input v-model="username" required class="mt-2 w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-brand-500 focus:ring-4 focus:ring-brand-100" /></label><label class="mt-4 block text-sm font-medium text-slate-700">Password<input v-model="password" required type="password" class="mt-2 w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-brand-500 focus:ring-4 focus:ring-brand-100" /></label><button :disabled="loading" class="mt-7 w-full rounded-xl bg-brand-600 px-4 py-3 font-semibold text-white transition hover:bg-brand-700 disabled:cursor-wait disabled:opacity-60">{{ loading ? 'Signing in...' : 'Sign in' }}</button></form></div></main>
</template>
