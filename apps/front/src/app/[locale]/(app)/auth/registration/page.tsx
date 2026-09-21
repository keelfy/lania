import AuthForm from '../components/auth-form'

// Kratos opens this page only when a provider left out data the account needs, e.g. a verified email.
// Signing in with a provider already creates the account, so the user never starts here.
export default function SignUpPage() {
  return <AuthForm type="registration" className="mx-auto flex-1" />
}
