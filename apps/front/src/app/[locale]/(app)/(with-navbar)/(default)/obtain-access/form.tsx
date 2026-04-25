import z from 'zod'
import { checkUsername } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'

export const MAX_PROFILES_PER_USER = 2

const usernameSchema = z
  .string()
  .min(3, { message: 'validation.username.minLength' })
  .max(16, { message: 'validation.username.maxLength' })
  .regex(/^[a-zA-Z0-9_]+$/, {
    message: 'validation.username.regex',
  })
  .superRefine(async (value, ctx) => {
    // Skip async check for obviously invalid or empty values to avoid noise
    if (!value || value.length < 3) return
    try {
      const res = await checkUsername(clientApiFetcher, value)
      if (res.some((r) => r.status === 'taken')) {
        ctx.addIssue({ code: 'custom', message: 'validation.username.taken' })
      } else if (res.some((r) => r.status === 'owned_by_you' && r.hasAccess)) {
        ctx.addIssue({
          code: 'custom',
          message: 'validation.username.already_has_access',
        })
      }
      // available or owned_by_you without access => valid
    } catch {
      ctx.addIssue({
        code: 'custom',
        message: 'validation.username.failed_to_check',
      })
    }
  })

export const obtainAccessFormSchema = z.object({
  username: usernameSchema,
})
