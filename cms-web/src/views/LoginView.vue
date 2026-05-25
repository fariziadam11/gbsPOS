<script setup lang="ts">
import { ref } from 'vue'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useLogin } from '../composables/useAuth'
import { getErrorMessage } from '../api/client'

const username = ref('')
const password = ref('')
const errorMsg = ref('')

const { mutate: doLogin, isPending } = useLogin()

function handleLogin() {
  errorMsg.value = ''
  if (!username.value || !password.value) {
    errorMsg.value = 'Please enter both username and password.'
    return
  }
  doLogin(
    { username: username.value, password: password.value },
    {
      onError: (err) => {
        errorMsg.value = getErrorMessage(err)
      },
    }
  )
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
        <form @submit.prevent="handleLogin" class="login-form">
          <div class="field">
            <label for="username">Username</label>
            <InputText
              id="username"
              v-model="username"
              placeholder="Enter username"
              :disabled="isPending"
              autocomplete="username"
              style="width: 100%"
            />
          </div>
          <div class="field">
            <label for="password">Password</label>
            <Password
              id="password"
              v-model="password"
              placeholder="Enter password"
              :feedback="false"
              :disabled="isPending"
              toggleMask
              style="width: 100%"
              :inputStyle="{ width: '100%' }"
            />
          </div>
          <Message
            v-if="errorMsg"
            severity="error"
            :closable="false"
            style="width: 100%"
          >
            {{ errorMsg }}
          </Message>
          <Button
            type="submit"
            label="Sign In"
            icon="pi pi-sign-in"
            style="width: 100%"
            :loading="isPending"
          />
        </form>
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

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field label {
  font-size: 14px;
  font-weight: 500;
  color: var(--p-text-secondary-color);
}
</style>
