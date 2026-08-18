export interface Product {
  id: number
  name: string
  price: number
  category: string
  imageUrl?: string
  stockQuantity: number
}

export interface CartItem extends Product {
  qty: number
}

export interface CardPayment {
  paymentId: string
  orderId: string
  amount: number
  status: string
  transactionId?: string
  failureReason?: string
}

export interface Order {
  id: string
  items: Array<{ productName: string; productPrice: number; qty: number; subtotal: number }>
  subtotal: number
  tax: number
  total: number
  paymentMethod: string
  timestamp: number
  transactionId?: string
  approvalCode?: string
  maskedAccount?: string
  isVoided?: boolean
}
