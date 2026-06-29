import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import { createPinia } from 'pinia'
import { VueQueryPlugin } from '@tanstack/vue-query'
import router from './router'
import App from './App.vue'
import { useAuthStore } from './stores/auth'
import 'primeicons/primeicons.css'
import 'primeflex/primeflex.css'

const savedTheme = localStorage.getItem('gbs-theme')
const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
const isDark = savedTheme === 'dark' || (savedTheme !== 'light' && systemDark)
if (isDark) {
  document.documentElement.classList.add('dark-mode')
}

const app = createApp(App)

app.use(createPinia())

const authStore = useAuthStore()
authStore.init()

app.use(router)
app.use(VueQueryPlugin)
app.use(ToastService)
app.use(ConfirmationService)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      prefix: 'p',
      darkModeSelector: '.dark-mode',
      cssLayer: false,
    },
  },
  ripple: true,
  inputVariant: 'filled',
})

app.mount('#app')
