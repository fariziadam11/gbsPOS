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
  <div class="login-page">
    <Card class="login-card">
      <template #title>
        <div class="login-title">
          <i class="pi pi-play-circle"></i>
          <span>GBS CMS</span>
        </div>
      </template>
      <template #subtitle>Sign in to manage ads</template>
      <template #content>
        <div class="login-form">
          <Message
            v-if="errorMsg"
            severity="error"
            :closable="false"
            style="width: 100%"
          >
            {{ errorMsg }}
          </Message>
          <Button
            type="button"
            label="Sign in with Keycloak"
            icon="pi pi-sign-in"
            style="width: 100%"
            :loading="isPending"
            @click="handleLogin"
          />
        </div>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--p-surface-50);
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 420px;
}

.login-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 600;
  color: var(--p-primary-color);
}

.login-title i {
  font-size: 32px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
