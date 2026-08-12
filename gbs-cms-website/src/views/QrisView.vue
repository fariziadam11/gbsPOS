<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { z } from 'zod'
import { getErrorMessage, posApi, requestData } from '../lib/api'

const orderId = ref('')
const submittedId = ref('')
const statusSchema = z.record(z.string(), z.unknown())
const status = useQuery({ queryKey: computed(() => ['qris', 'status', submittedId.value]), queryFn: () => requestData(posApi, { method: 'GET', url: `/qris/payments/${submittedId.value}/status` }, statusSchema), enabled: computed(() => Boolean(submittedId.value)) })
const statusError = computed(() => status.isError.value)
const statusPending = computed(() => status.isPending.value)
const statusErrorMessage = computed(() => getErrorMessage(status.error.value))
function lookup() { submittedId.value = orderId.value.trim() }
</script>

<template>
  <div class="space-y-6"><div><p class="text-sm text-slate-500">Payment operations</p><h2 class="mt-1 text-3xl font-bold text-slate-900">QRIS Payments</h2></div><section class="max-w-2xl rounded-2xl border border-slate-200 bg-white p-6"><h3 class="font-bold text-slate-900">Check payment status</h3><p class="mt-1 text-sm text-slate-500">Endpoint POS menyediakan lookup status berdasarkan order ID, bukan daftar transaksi.</p><form class="mt-5 flex flex-col gap-3 sm:flex-row" @submit.prevent="lookup"><input v-model="orderId" required class="min-w-0 flex-1 rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-brand-500" placeholder="Order ID" /><button class="rounded-xl bg-brand-600 px-5 py-3 font-semibold text-white hover:bg-brand-700">Check status</button></form><div v-if="statusError" class="mt-5 rounded-xl bg-red-50 p-4 text-sm text-red-700">{{ statusErrorMessage }}</div><div v-if="statusPending" class="mt-5 text-sm text-slate-500">Checking payment...</div><pre v-if="status.data.value" class="mt-5 overflow-auto rounded-xl bg-slate-950 p-4 text-xs text-emerald-300">{{ JSON.stringify(status.data.value, null, 2) }}</pre></section></div>
</template>
