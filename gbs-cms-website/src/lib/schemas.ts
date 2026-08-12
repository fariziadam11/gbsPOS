import { z } from 'zod'

export const productSchema = z.object({ id: z.number(), name: z.string(), price: z.number(), category: z.string().optional().default(''), storeType: z.string().optional().default(''), stockQuantity: z.number().optional().default(0), finalPrice: z.number().optional() }).passthrough()
export const orderSchema = z.object({ id: z.string(), total: z.number(), paymentMethod: z.string().optional().default(''), timestamp: z.number().optional(), storeType: z.string().optional().default(''), isVoided: z.boolean().optional(), isSettled: z.boolean().optional() }).passthrough()
export const customerSchema = z.object({ id: z.number(), name: z.string().optional().default(''), phone: z.string().optional().default(''), email: z.string().optional().default('') }).passthrough()
export const userSchema = z.object({ id: z.number(), username: z.string(), name: z.string(), role: z.string(), gender: z.string().optional().default('') }).passthrough()
export const discountSchema = z.object({ id: z.number(), name: z.string(), type: z.string().optional().default(''), value: z.number().optional().default(0), status: z.string().optional().default('') }).passthrough()
export const adSchema = z.object({ id: z.number(), name: z.string(), filename: z.string().optional().default(''), isActive: z.boolean().optional().default(false), playlistOrder: z.number().optional().default(0), storeTypes: z.array(z.string()).optional().default([]) }).passthrough()
export const summarySchema = z.object({ totalOrders: z.number(), totalRevenue: z.number(), avgOrderValue: z.number(), cashTotal: z.number(), cardTotal: z.number(), qrisTotal: z.number(), voidedCount: z.number() })
export const revenueSchema = z.object({ date: z.string(), revenue: z.number(), orders: z.number() })
export const settingsSchema = z.object({ settings: z.record(z.string(), z.string()) })
export const paginationSchema = z.object({ page: z.number(), limit: z.number(), total: z.number(), totalPages: z.number() })
