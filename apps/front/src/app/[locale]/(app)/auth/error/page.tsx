'use client'

import LoadingSpinner from '@/components/ui/loading-spinner'
import ory from '@/lib/ory'
import { FlowError } from '@ory/client-fetch'
import { parseAsString, useQueryState } from 'nuqs'
import React from 'react'

export default function ErrorPage() {
  const [id] = useQueryState('id', parseAsString.withDefault(''))
  const [flowError, setFlowError] = React.useState<FlowError>()
  const [isErrorLoading, startTransition] = React.useTransition()

  React.useEffect(() => {
    startTransition(async () => {
      const error = await ory.getFlowError({ id })
      setFlowError(error)
    })
  }, [id])

  return (
    <div className="flex h-screen w-screen flex-col items-center justify-center">
      <h1 className="text-2xl font-bold">
        We couldn&apos;t process your authentication request
      </h1>
      {isErrorLoading ? (
        <LoadingSpinner className="text-center" />
      ) : (
        <pre className="text-sm text-gray-500">
          {JSON.stringify(flowError, null, 2)}
        </pre>
      )}
    </div>
  )
}
