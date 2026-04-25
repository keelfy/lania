import { Currency } from '@/lib/currency'

export enum PaymentMethod {
  DonationAlerts = 'donation-alerts',
  Freekassa = 'freekassa',
  EasyDonate = 'easy-donate',
}

export type CreateOrderReq = {
  paymentMethod: PaymentMethod
  products: CreateOrderProductReq[]
}

export type CreateOrderProductReq = {
  id: string
  profileId: string
}

export type CreateOrderRes = {
  paymentUrl: string
}

export type OrderStatus = 'created' | 'processing' | 'completed' | 'failed'

export type OrderItem = {
  id: string
  productId: string
  profileId: string
  seasonId: string
  quantity: number
  amounts: OrderAmount[]
}

export type OrderAmount = {
  currency: Currency
  amount: number
}

export type Order = {
  id: string
  status: OrderStatus
  amounts: OrderAmount[]
  items: OrderItem[]
  createdAt: string
}

export type PurchasedProduct = {
  productId: string
  profileId: string
  seasonId: string
}
