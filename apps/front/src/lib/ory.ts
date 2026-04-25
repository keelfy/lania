import { FrontendApi, Configuration } from '@ory/client-fetch'

const ory = new FrontendApi(
  new Configuration({
    basePath: process.env.NEXT_PUBLIC_ORY_SDK_URL || 'http://localhost:4000',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    credentials: 'include',
  }),
)

export const fetchOry = async (path: string, options?: RequestInit) => {
  const response = await fetch(
    `${process.env.NEXT_PUBLIC_ORY_SDK_URL}${path}`,
    {
      ...options,
      credentials: 'include',
    },
  )
  return response
}

export default ory
