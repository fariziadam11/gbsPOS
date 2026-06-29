<script setup lang="ts">
import { ref } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const errorMsg = ref('')
const isPending = ref(false)

async function handleLogin() {
  errorMsg.value = ''
  isPending.value = true
  try {
    await authStore.login()
  } catch (err) {
    isPending.value = false
    errorMsg.value = err instanceof Error ? err.message : 'Failed to start login'
  }
}
</script>

<template>
  <div class="min-h-screen flex align-items-center justify-content-center surface-50 p-4">
    <Card class="w-full max-w-30rem">
      <template #title>
        <div class="flex align-items-center gap-2 text-2xl font-semibold text-primary">
          <i class="pi pi-play-circle text-3xl"></i>
          <span>GBS CMS</span>
        </div>
      </template>
      <template #subtitle>Sign in to manage ads</template>
      <template #content>
        <div class="flex flex-column gap-3">
          <Message
            v-if="errorMsg"
            severity="error"
            :closable="false"
            class="w-full"
          >
            {{ errorMsg }}
          </Message>
          <Button
            type="button"
            label="Sign in with Keycloak"
            icon="pi pi-sign-in"
            class="w-full"
            :loading="isPending"
            @click="handleLogin"
          />
        </div>
      </template>
    </Card>
  </div>
</template>
