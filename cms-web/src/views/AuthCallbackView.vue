<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const error = ref('')

onMounted(async () => {
  try {
    await authStore.handleLoginCallback(window.location.href)
    const redirect = route.query.redirect as string | undefined
    router.replace(redirect || '/')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Login failed'
  }
})
</script>

<template>
  <div class="min-h-screen flex align-items-center justify-content-center surface-50 p-4">
    <div v-if="error" class="flex flex-column align-items-center gap-3 text-center">
      <h2 class="text-red-500 m-0">Authentication failed</h2>
      <p>{{ error }}</p>
      <router-link to="/login" class="text-primary underline">Back to login</router-link>
    </div>
    <div v-else class="flex flex-column align-items-center gap-3 text-center">
      <i class="pi pi-spin pi-spinner" style="font-size: 2rem"></i>
      <p>Completing sign in...</p>
    </div>
  </div>
</template>
