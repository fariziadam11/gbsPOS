import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { login } from '../api/auth'
import { useAuthStore } from '../stores/auth'

export function useLogin() {
  const router = useRouter()
  const authStore = useAuthStore()

  return useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      if (data.success) {
        authStore.setToken(data.data.token)
        authStore.setUser(data.data.user)
        router.push('/')
      }
    },
  })
}
