<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { cancelCardPayment, createOrder, getOrders, getProducts, initCardPayment, login, websocketUrl } from './api'
import type { CartItem, CardPayment, Order, Product } from './types'

const token = ref(localStorage.getItem('gbs_pos_token'))
const username = ref('cashier')
const password = ref('cashier123')
const error = ref('')
const products = ref<Product[]>([])
const cart = ref<CartItem[]>([])
const payment = ref<CardPayment | null>(null)
const currentOrder = ref<Order | null>(null)
const orders = ref<Order[]>([])
const screen = ref<'checkout' | 'receipt' | 'history'>('checkout')
const historyLoading = ref(false)
const search = ref('')
const activeCategory = ref('Semua')
const connection = ref('DISCONNECTED')
const terminalId = import.meta.env.VITE_TERMINAL_ID ?? 'POS-001'
const deviceId = import.meta.env.VITE_COMPANION_DEVICE_ID ?? 'HP-KASIR-01'
let socket: WebSocket | undefined
let reconnectTimer: ReturnType<typeof setTimeout> | undefined

const total = computed(() => cart.value.reduce((sum, item) => sum + item.price * item.qty, 0))
const categories = computed(() => [...new Set(products.value.map((product) => product.category))])
const filteredProducts = computed(() => products.value.filter((product) => {
  const matchesCategory = activeCategory.value === 'Semua' || product.category === activeCategory.value
  return matchesCategory && product.name.toLowerCase().includes(search.value.toLowerCase())
}))

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
      if (message.status === 'SUCCESS') {
        cart.value = []
      }
      if (message.status === 'SUCCESS' && currentOrder.value?.id === message.order_id) screen.value = 'receipt'
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

function increment(item: CartItem) {
  item.qty += 1
}

async function payByCard() {
  error.value = ''
  const orderId = `WEB-${Date.now()}`
  try {
    currentOrder.value = await createOrder({
      id: orderId,
      terminalId,
      subtotal: total.value,
      total: total.value,
      items: cart.value.map((item) => ({ productId: item.id, productName: item.name, productPrice: item.price, qty: item.qty, subtotal: item.price * item.qty })),
    })
    payment.value = await initCardPayment(orderId, total.value, terminalId, deviceId)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'Gagal membuat pembayaran kartu.'
  }
}

async function showHistory() {
  screen.value = 'history'
  historyLoading.value = true
  try {
    orders.value = await getOrders(terminalId)
  } catch {
    error.value = 'Gagal mengambil riwayat transaksi.'
  } finally {
    historyLoading.value = false
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

function showReceipt(order: Order) {
  currentOrder.value = order
  screen.value = 'receipt'
}

// Transaksi kartu yang gagal = status payment FAILED/CANCELLED/EXPIRED, atau order CARD tanpa transactionId
function isFailedOrder(order: Order) {
  const livePayment = payment.value
  if (livePayment && livePayment.orderId === order.id && (livePayment.status === 'FAILED' || livePayment.status === 'CANCELLED' || livePayment.status === 'EXPIRED')) return true
  if (livePayment && livePayment.orderId === order.id && livePayment.status === 'SUCCESS') return false
  return order.paymentMethod === 'CARD' && !order.transactionId
}

function orderStatus(order: Order) {
  return isFailedOrder(order) ? 'Gagal' : 'Berhasil'
}

function newTransaction() {
  screen.value = 'checkout'
  payment.value = null
  currentOrder.value = null
  cart.value = []
  search.value = ''
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
      <div class="brand-mark">GBS</div>
      <p class="eyebrow">POINT OF SALE / WEB</p>
      <h1>Mulai transaksi.</h1>
      <p class="login-copy">Masuk untuk membuka terminal kasir {{ terminalId }}.</p>
      <label>Username<input v-model="username" autocomplete="username" /></label>
      <label>Password<input v-model="password" type="password" autocomplete="current-password" /></label>
      <p v-if="error" class="error">{{ error }}</p>
      <button class="primary-button">Masuk ke POS</button>
    </form>
  </main>
  <main v-else class="pos-shell">
    <header class="topbar">
      <div class="topbar-title"><p class="eyebrow">TERMINAL {{ terminalId }}</p><h1>{{ screen === 'checkout' ? 'Checkout' : screen === 'receipt' ? 'Receipt' : 'Riwayat transaksi' }}</h1></div>
      <nav class="topbar-actions" aria-label="Navigasi POS">
        <button :class="['nav-button', { active: screen === 'checkout' }]" @click="screen = 'checkout'">Kasir</button>
        <button :class="['nav-button', { active: screen === 'history' }]" @click="showHistory">Riwayat</button>
        <span :class="['connection', connection.toLowerCase()]"><i />{{ connection }}</span>
        <button class="quiet" @click="logout">Keluar</button>
      </nav>
    </header>
    <p v-if="error" class="error banner">{{ error }}</p>

    <section v-if="screen === 'checkout'" class="workspace">
      <div class="catalog">
        <div class="section-heading"><div><p class="eyebrow">CATALOG</p><h2>Produk</h2></div><span class="count-label">{{ filteredProducts.length }} item</span></div>
        <div class="catalog-tools"><label class="search-field"><span aria-hidden="true">⌕</span><input v-model="search" placeholder="Cari produk..." aria-label="Cari produk" /></label><div class="category-tabs" role="tablist"><button :class="['category-tab', { active: activeCategory === 'Semua' }]" @click="activeCategory = 'Semua'">Semua</button><button v-for="category in categories" :key="category" :class="['category-tab', { active: activeCategory === category }]" @click="activeCategory = category">{{ category }}</button></div></div>
        <div v-if="!filteredProducts.length" class="empty catalog-empty">Produk tidak ditemukan.</div>
        <div v-else class="product-grid"><button v-for="product in filteredProducts" :key="product.id" class="product" @click="add(product)"><span class="product-category">{{ product.category }}</span><strong>{{ product.name }}</strong><span class="product-price">Rp {{ product.price.toLocaleString('id-ID') }}</span></button></div>
      </div>
      <aside class="cart-panel">
        <div class="section-heading"><div><p class="eyebrow">CURRENT SALE</p><h2>Keranjang</h2></div><span class="count-label">{{ cart.length }} item</span></div>
        <div v-if="!cart.length" class="empty cart-empty"><span class="empty-icon">+</span><strong>Keranjang masih kosong</strong><small>Pilih produk di sebelah kiri untuk mulai.</small></div>
        <div v-for="item in cart" :key="item.id" class="cart-item"><div class="cart-item-info"><strong>{{ item.name }}</strong><small>Rp {{ item.price.toLocaleString('id-ID') }}</small></div><div class="quantity-control"><button aria-label="Kurangi jumlah" @click="remove(item)">−</button><span>{{ item.qty }}</span><button aria-label="Tambah jumlah" @click="increment(item)">+</button></div></div>
        <div class="total-row"><span>Total</span><strong>Rp {{ total.toLocaleString('id-ID') }}</strong></div>
        <button class="primary-button pay" :disabled="!cart.length || payment?.status === 'WAITING_FOR_CARD' || payment?.status === 'PROCESSING'" @click="payByCard"><span>Bayar dengan Kartu</span><span>→</span></button>
        <div v-if="payment" class="payment-status"><div class="status-heading"><span :class="['status-dot', payment.status.toLowerCase()]" /><span class="eyebrow">PAYMENT {{ payment.status }}</span></div><strong v-if="payment.status === 'WAITING_FOR_CARD'">Silakan tap kartu di companion app</strong><strong v-else-if="payment.status === 'SUCCESS'">Pembayaran berhasil</strong><strong v-else>{{ payment.failureReason || payment.status }}</strong><button v-if="payment.status === 'WAITING_FOR_CARD'" class="quiet" @click="cancelPayment">Batalkan pembayaran</button></div>
      </aside>
    </section>

    <section v-else-if="screen === 'receipt'" class="single-panel">
      <div v-if="currentOrder" class="receipt-card">
        <template v-if="isFailedOrder(currentOrder)">
          <div class="fail-badge">✕</div><p class="eyebrow">TRANSACTION FAILED</p><h2>Pembayaran gagal</h2>
        </template>
        <template v-else>
          <div class="success-badge">✓</div><p class="eyebrow">TRANSACTION COMPLETE</p><h2>Pembayaran berhasil</h2>
        </template>
        <p class="receipt-id">{{ currentOrder.id }}</p><div class="receipt-items"><div v-for="item in currentOrder.items" :key="`${currentOrder.id}-${item.productName}`" class="cart-item"><div class="cart-item-info"><strong>{{ item.productName }}</strong><small>{{ item.qty }} × Rp {{ item.productPrice.toLocaleString('id-ID') }}</small></div><strong>Rp {{ item.subtotal.toLocaleString('id-ID') }}</strong></div></div><div class="total-row"><span>Total</span><strong>Rp {{ currentOrder.total.toLocaleString('id-ID') }}</strong></div><p class="receipt-meta">{{ currentOrder.paymentMethod }} · {{ currentOrder.transactionId || payment?.transactionId || '-' }}</p><button class="primary-button pay" @click="newTransaction">Transaksi Baru <span>→</span></button></div><div v-else class="empty">Receipt belum tersedia.</div>
    </section>
    <section v-else class="history-panel"><div class="section-heading"><div><p class="eyebrow">TERMINAL {{ terminalId }}</p><h2>Transaksi terakhir</h2></div><button class="quiet" @click="showHistory">Refresh</button></div><div v-if="historyLoading" class="empty">Memuat transaksi...</div><div v-else-if="!orders.length" class="empty">Belum ada transaksi.</div><template v-else><button v-for="order in orders" :key="order.id" class="history-row" @click="showReceipt(order)"><span><strong>{{ order.id }}</strong><small>{{ new Date(order.timestamp).toLocaleString('id-ID') }} · {{ order.paymentMethod }} · <span :class="['status-label', { failed: isFailedOrder(order) }]">{{ orderStatus(order) }}</span></small></span><strong>Rp {{ order.total.toLocaleString('id-ID') }}</strong></button></template></section>
  </main>
</template>
