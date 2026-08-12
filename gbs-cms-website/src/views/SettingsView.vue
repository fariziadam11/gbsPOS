<script setup lang="ts">
import { reactive, watch } from 'vue'
import { computed } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { z } from 'zod'
import { cmsApi, getErrorMessage, requestData } from '../lib/api'
import { settingsSchema } from '../lib/schemas'

const queryClient = useQueryClient()
const settings = reactive<Record<string, string>>({})
const query = useQuery({ queryKey: ['settings'], queryFn: () => requestData(cmsApi, { method: 'GET', url: '/settings' }, settingsSchema) })
watch(() => query.data.value, (value) => { if (value) { Object.keys(settings).forEach((key) => delete settings[key]); Object.assign(settings, value.settings) } }, { immediate: true })
const save = useMutation({ mutationFn: () => requestData(cmsApi, { method: 'PUT', url: '/settings', data: { settings: z.record(z.string(), z.string()).parse(settings) } }, settingsSchema), onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['settings'] }) })
const isSaving = computed(() => save.isPending.value)
const isLoading = computed(() => query.isPending.value)
const hasQueryError = computed(() => query.isError.value)
const isSaved = computed(() => save.isSuccess.value)
const hasSaveError = computed(() => save.isError.value)
const queryErrorMessage = computed(() => getErrorMessage(query.error.value))
const saveErrorMessage = computed(() => getErrorMessage(save.error.value))
function submit() { save.mutate() }
</script>

<template>
  <div class="space-y-6"><div><p class="text-sm text-slate-500">System configuration</p><h2 class="mt-1 text-3xl font-bold text-slate-900">Settings</h2></div><div v-if="hasQueryError" class="rounded-2xl bg-red-50 p-4 text-sm text-red-700">{{ queryErrorMessage }}</div><form class="max-w-3xl space-y-5 rounded-2xl border border-slate-200 bg-white p-6" @submit.prevent="submit"><div v-if="isLoading" class="text-sm text-slate-500">Loading settings...</div><label v-for="key in Object.keys(settings)" :key="key" class="block text-sm font-semibold text-slate-700">{{ key }}<input v-model="settings[key]" class="mt-2 w-full rounded-xl border border-slate-200 px-4 py-3 font-normal outline-none focus:border-brand-500 focus:ring-4 focus:ring-brand-100" /></label><button :disabled="isSaving || isLoading" class="rounded-xl bg-brand-600 px-5 py-3 font-semibold text-white disabled:opacity-50">{{ isSaving ? 'Saving...' : 'Save settings' }}</button><p v-if="isSaved" class="text-sm text-brand-600">Settings saved.</p><p v-if="hasSaveError" class="text-sm text-red-600">{{ saveErrorMessage }}</p></form></div>
</template>
