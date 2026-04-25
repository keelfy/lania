import React from 'react'
import { toast } from 'sonner'

export const errorToast = (title: string | React.ReactNode, error: unknown) =>
  toast.error(title, {
    description: error instanceof Error ? error.message : 'Попробуйте позже',
  })
