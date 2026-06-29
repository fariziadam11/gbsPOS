import { ref, watch, onMounted } from 'vue'

type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'gbs-theme'
const CLASS_NAME = 'dark-mode'

function getSystemDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyDark(isDark: boolean) {
  if (isDark) {
    document.documentElement.classList.add(CLASS_NAME)
  } else {
    document.documentElement.classList.remove(CLASS_NAME)
  }
}

function loadSavedMode(): ThemeMode {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'light' || saved === 'dark' || saved === 'system') {
    return saved
  }
  return 'system'
}

const mode = ref<ThemeMode>('system')
const isDark = ref(false)

export function useTheme() {
  function sync() {
    if (mode.value === 'system') {
      isDark.value = getSystemDark()
    } else {
      isDark.value = mode.value === 'dark'
    }
    applyDark(isDark.value)
  }

  function setMode(newMode: ThemeMode) {
    mode.value = newMode
    localStorage.setItem(STORAGE_KEY, newMode)
    sync()
  }

  function toggle() {
    setMode(isDark.value ? 'light' : 'dark')
  }

  onMounted(() => {
    mode.value = loadSavedMode()
    sync()

    const media = window.matchMedia('(prefers-color-scheme: dark)')
    media.addEventListener('change', () => {
      if (mode.value === 'system') {
        sync()
      }
    })
  })

  watch(mode, sync)

  return {
    mode,
    isDark,
    setMode,
    toggle,
  }
}
