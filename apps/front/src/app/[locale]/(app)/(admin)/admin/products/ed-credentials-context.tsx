'use client'

import React from 'react'

type EdCredentials = { userAuth: string }

type EdCredentialsStore = {
  hasCredentials: boolean
  getCredentials: () => EdCredentials | undefined
  setCredentials: (value: EdCredentials) => void
}

const EdCredentialsContext = React.createContext<EdCredentialsStore | null>(
  null,
)

// Holds the admin's EasyDonate user_auth cookie in memory for the lifetime of the
// products page, so several positions can be created in a row without retyping it. The value
// lives in a ref, not state: nothing here is written to web storage or sent anywhere but the one
// create-product request, and typing in the dialog must not re-render the rest of the page.
export function EdCredentialsProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const credentialsRef = React.useRef<EdCredentials>(undefined)
  const [hasCredentials, setHasCredentials] = React.useState(false)

  const getCredentials = React.useCallback(() => credentialsRef.current, [])
  const setCredentials = React.useCallback((value: EdCredentials) => {
    credentialsRef.current = value
    setHasCredentials(true)
  }, [])

  const value = React.useMemo<EdCredentialsStore>(
    () => ({ hasCredentials, getCredentials, setCredentials }),
    [hasCredentials, getCredentials, setCredentials],
  )

  return (
    <EdCredentialsContext.Provider value={value}>
      {children}
    </EdCredentialsContext.Provider>
  )
}

export function useEdCredentials(): EdCredentialsStore {
  const context = React.useContext(EdCredentialsContext)
  if (!context) {
    throw new Error(
      'useEdCredentials must be used within an EdCredentialsProvider',
    )
  }
  return context
}
