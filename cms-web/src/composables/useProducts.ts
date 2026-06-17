import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import type { Ref } from 'vue'
import { getProducts, getProduct, createProduct, updateProduct, deleteProduct, importProducts } from '../api/products'
import type { UpdateProductRequest } from '../types/api'

export function useProducts(storeType?: Ref<string | undefined>) {
  return useQuery({
    queryKey: computed(() => ['products', storeType?.value]),
    queryFn: () => getProducts(storeType?.value),
    select: (res) => (res.success ? res.data : null),
  })
}

export function useProduct(id: Ref<number>) {
  return useQuery({
    queryKey: ['product', id.value],
    queryFn: () => getProduct(id.value),
    select: (res) => (res.success ? res.data : null),
    enabled: () => id.value > 0,
  })
}

export function useCreateProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createProduct,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useUpdateProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateProductRequest }) => updateProduct(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useDeleteProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteProduct,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useImportProducts() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ file, storeType }: { file: File; storeType?: string }) => importProducts(file, storeType),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })
}
