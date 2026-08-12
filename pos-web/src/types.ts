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
