<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { cancelCardPayment, createOrder, getProducts, initCardPayment, login, websocketUrl } from './api'
import type { CartItem, CardPayment, Product } from './types'

const token = ref(localStorage.getItem('gbs_pos_token'))
const username = ref('cashier')
const password = ref('cashier123')
const error = ref('')
const products = ref<Product[]>([])
const cart = ref<CartItem[]>([])
const payment = ref<CardPayment | null>(null)
const connection = ref('DISCONNECTED')
const terminalId = import.meta.env.VITE_TERMINAL_ID ?? 'POS-001'
const deviceId = import.meta.env.VITE_COMPANION_DEVICE_ID ?? 'HP-KASIR-01'
let socket: WebSocket | undefined
let reconnectTimer: ReturnType<typeof setTimeout> | undefined

const total = computed(() => cart.value.reduce((sum, item) => sum + item.price * item.qty, 0))
const categories = computed(() => [...new Set(products.value.map((product) => product.category))])

async function signIn() {
  try {
    const response = await login(username.value, password.value)
    token.value = response.data.token
    localStorage.setItem('gbs_pos_token', token.value)
    await load()
  } catch {
    error.value = 'Login gagal. Periksa username dan password.'
  }
}

async function load() {
  try {
    products.value = await getProducts()
    connect()
  } catch {
    error.value = 'Tidak dapat mengambil produk dari backend.'
  }
}

function connect() {
  socket?.close()
  socket = new WebSocket(websocketUrl(terminalId))
  socket.onopen = () => { connection.value = 'CONNECTED' }
  socket.onclose = () => {
    connection.value = 'DISCONNECTED'
    if (token.value) reconnectTimer = setTimeout(connect, 3000)
  }
  socket.onmessage = (event) => {
    const message = JSON.parse(event.data) as {
      type: string
      payment_id?: string
      order_id?: string
      amount?: number
      status?: string
      transaction_id?: string
      failure_reason?: string
    }
    if (message.type === 'PAYMENT_STATUS') {
      payment.value = {
        paymentId: message.payment_id ?? '',
        orderId: message.order_id ?? '',
        amount: message.amount ?? 0,
        status: message.status ?? 'FAILED',
        transactionId: message.transaction_id,
        failureReason: message.failure_reason,
      }
      if (message.status === 'SUCCESS') cart.value = []
    }
  }
}

function add(product: Product) {
  const item = cart.value.find((candidate) => candidate.id === product.id)
  if (item) item.qty += 1
  else cart.value.push({ ...product, qty: 1 })
}

function remove(item: CartItem) {
  if (item.qty > 1) item.qty -= 1
  else cart.value = cart.value.filter((candidate) => candidate.id !== item.id)
}

async function payByCard() {
  error.value = ''
  const orderId = `WEB-${Date.now()}`
  try {
    await createOrder({
      id: orderId,
      terminalId,
      subtotal: total.value,
      total: total.value,
      items: cart.value.map((item) => ({ productId: item.id, productName: item.name, productPrice: item.price, qty: item.qty, subtotal: item.price * item.qty })),
    })
    payment.value = await initCardPayment(orderId, total.value, terminalId, deviceId)
  } catch {
    error.value = 'Gagal membuat pembayaran kartu.'
  }
}

async function cancelPayment() {
  if (!payment.value) return
  try {
    payment.value = (await cancelCardPayment(payment.value.paymentId)).data
  } catch {
    error.value = 'Gagal membatalkan pembayaran.'
  }
}

function logout() {
  localStorage.removeItem('gbs_pos_token')
  token.value = null
  socket?.close()
  if (reconnectTimer) clearTimeout(reconnectTimer)
}

onMounted(() => { if (token.value) load() })
onUnmounted(() => {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  socket?.close()
})
</script>

<template>
  <main v-if="!token" class="login-shell">
    <form class="login-card" @submit.prevent="signIn">
      <p class="eyebrow">GBS POS / WEB</p><h1>Mulai transaksi</h1>
      <label>Username<input v-model="username" autocomplete="username" /></label>
      <label>Password<input v-model="password" type="password" autocomplete="current-password" /></label>
      <p v-if="error" class="error">{{ error }}</p><button>Masuk ke POS</button>
    </form>
  </main>
  <main v-else class="pos-shell">
    <header><div><p class="eyebrow">TERMINAL {{ terminalId }}</p><h1>Checkout</h1></div><div class="header-actions"><span :class="['connection', connection.toLowerCase()]">{{ connection }}</span><button class="quiet" @click="logout">Keluar</button></div></header>
    <p v-if="error" class="error banner">{{ error }}</p>
    <section class="workspace">
      <div class="catalog"><div class="section-heading"><div><p class="eyebrow">CATALOG</p><h2>Produk</h2></div><span>{{ products.length }} item</span></div><div v-for="category in categories" :key="category" class="category"><h3>{{ category }}</h3><div class="product-grid"><button v-for="product in products.filter((item) => item.category === category)" :key="product.id" class="product" @click="add(product)"><strong>{{ product.name }}</strong><span>Rp {{ product.price.toLocaleString('id-ID') }}</span></button></div></div></div>
     <aside class="cart-panel"><div class="section-heading"><div><p class="eyebrow">CURRENT SALE</p><h2>Keranjang</h2></div><span>{{ cart.length }} item</span></div><div v-if="!cart.length" class="empty">Pilih produk untuk memulai transaksi.</div><div v-for="item in cart" :key="item.id" class="cart-item"><div><strong>{{ item.name }}</strong><small>{{ item.qty }} × Rp {{ item.price.toLocaleString('id-ID') }}</small></div><button class="remove" @click="remove(item)">−</button></div><div class="total-row"><span>Total</span><strong>Rp {{ total.toLocaleString('id-ID') }}</strong></div><button class="pay" :disabled="!cart.length || payment?.status === 'WAITING_FOR_CARD' || payment?.status === 'PROCESSING'" @click="payByCard">Bayar dengan Kartu</button><div v-if="payment" class="payment-status"><span class="eyebrow">PAYMENT {{ payment.status }}</span><strong v-if="payment.status === 'WAITING_FOR_CARD'">Silakan tap kartu di companion app</strong><strong v-else-if="payment.status === 'SUCCESS'">Pembayaran berhasil</strong><strong v-else>{{ payment.failureReason || payment.status }}</strong><button v-if="payment.status === 'WAITING_FOR_CARD'" class="quiet" @click="cancelPayment">Batalkan pembayaran</button></div></aside>
    </section>
  </main>
</template>
