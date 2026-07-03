<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Password from 'primevue/password'
import { getErrorMessage } from '../api/client'
import { useAuthStore } from '../stores/auth'
import { authMode } from '../keycloak'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const errorMsg = ref('')
const isPending = ref(false)
const username = ref('')
const password = ref('')
const isBasicAuth = computed(() => authMode === 'basic')

async function redirectAfterLogin() {
  const redirect = Array.isArray(route.query.redirect)
    ? route.query.redirect[0]
    : route.query.redirect
  await router.replace(typeof redirect === 'string' && redirect ? redirect : '/')
}

async function handleKeycloakLogin() {
  errorMsg.value = ''
  isPending.value = true
  try {
    await authStore.login()
  } catch (err) {
    isPending.value = false
    errorMsg.value = err instanceof Error ? err.message : 'Failed to start login'
  }
}

async function handleBasicLogin() {
  errorMsg.value = ''
  if (!username.value.trim() || !password.value) {
    errorMsg.value = 'Username and password are required'
    return
  }

  isPending.value = true
  try {
    await authStore.login({
      username: username.value.trim(),
      password: password.value,
    })
    await redirectAfterLogin()
  } catch (err) {
    errorMsg.value = getErrorMessage(err)
  } finally {
    isPending.value = false
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
      <template #subtitle>
        {{ isBasicAuth ? 'Sign in with your CMS account' : 'Sign in to manage ads' }}
      </template>
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
          <form
            v-if="isBasicAuth"
            class="flex flex-column gap-3"
            @submit.prevent="handleBasicLogin"
          >
            <div class="flex flex-column gap-2">
              <label for="username" class="font-medium">Username</label>
              <InputText
                id="username"
                v-model="username"
                autocomplete="username"
                :disabled="isPending"
                autofocus
              />
            </div>
            <div class="flex flex-column gap-2">
              <label for="password" class="font-medium">Password</label>
              <Password
                id="password"
                v-model="password"
                autocomplete="current-password"
                class="w-full"
                input-class="w-full"
                :feedback="false"
                toggle-mask
                :disabled="isPending"
              />
            </div>
            <Button
              type="submit"
              label="Sign in"
              icon="pi pi-sign-in"
              class="w-full"
              :loading="isPending"
            />
          </form>
          <Button
            v-else
            type="button"
            label="Sign in with Keycloak"
            icon="pi pi-sign-in"
            class="w-full"
            :loading="isPending"
            @click="handleKeycloakLogin"
          />
        </div>
      </template>
    </Card>
  </div>
</template>
