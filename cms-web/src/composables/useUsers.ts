import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getUsers, createUser, updateUser, deleteUser } from '../api/users'
import type { UpdateUserRequest } from '../types/api'

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
    select: (res) => (res.success ? res.data : null),
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateUserRequest }) => updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })
}
