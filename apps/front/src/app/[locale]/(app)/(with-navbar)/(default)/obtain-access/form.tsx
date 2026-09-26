import z from 'zod'
import { checkUsername } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { UsernameCheck } from '@/models/profile'

export const MAX_PROFILES_PER_USER = 2

const CHECK_DEBOUNCE_MS = 400

type CheckResult = { username: string; checks: UsernameCheck[] }

let checkTimer: ReturnType<typeof setTimeout> | undefined
let checkWaiters: {
  resolve: (result: CheckResult) => void
  reject: (error: unknown) => void
}[] = []

// The form validates on every keystroke, and the check asks Mojang whether the
// nickname is premium. Only the last value typed is checked; earlier calls get
// its result, so the form never waits on a stale nickname.
function checkUsernameDebounced(
  username: string,
  seasonId?: string,
): Promise<CheckResult> {
  clearTimeout(checkTimer)
  return new Promise((resolve, reject) => {
    checkWaiters.push({ resolve, reject })
    checkTimer = setTimeout(() => {
      const waiters = checkWaiters
      checkWaiters = []
      checkUsername(clientApiFetcher, username, seasonId).then(
        (checks) => waiters.forEach((w) => w.resolve({ username, checks })),
        (error) => waiters.forEach((w) => w.reject(error)),
      )
    }, CHECK_DEBOUNCE_MS)
  })
}

// The username check depends on the season: a profile may have access to one season and not to another.
const createUsernameSchema = (
  seasonId: string | undefined,
  onChecked: (username: string, premium: boolean) => void,
) =>
  z
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
        const { username, checks: res } = await checkUsernameDebounced(
          value,
          seasonId,
        )
        onChecked(
          username,
          res.some((r) => r.premium),
        )
        if (res.some((r) => r.status === 'taken')) {
          ctx.addIssue({ code: 'custom', message: 'validation.username.taken' })
        } else if (
          res.some((r) => r.status === 'owned_by_you' && r.hasAccess)
        ) {
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

// onPremiumChecked tells the page which nickname was checked and whether a
// Mojang account has it: the server lets only the owner of such an account in,
// so the player has to acknowledge that before getting access.
export const createObtainAccessFormSchema = (
  seasonId?: string,
  onPremiumChecked?: (username: string, premium: boolean) => void,
) => {
  let premium = false
  return z
    .object({
      username: createUsernameSchema(seasonId, (username, isPremium) => {
        premium = isPremium
        onPremiumChecked?.(username, isPremium)
      }),
      premiumAcknowledged: z.boolean(),
    })
    .superRefine((values, ctx) => {
      if (premium && !values.premiumAcknowledged) {
        ctx.addIssue({
          code: 'custom',
          path: ['premiumAcknowledged'],
          message: 'validation.premium.required',
        })
      }
    })
}

export type ObtainAccessFormValues = z.infer<
  ReturnType<typeof createObtainAccessFormSchema>
>
