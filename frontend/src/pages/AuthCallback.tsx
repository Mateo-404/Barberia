import { useEffect, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { useGoogleCallback, useTOTPChallenge } from "@/api/auth"
import { writeSession } from "@/lib/auth-storage"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { errorMessage } from "@/lib/error-message"

export default function AuthCallback() {
  const { t: tc } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const googleCallback = useGoogleCallback()
  const challenge = useTOTPChallenge()
  const [error, setError] = useState<string | null>(null)
  const [pendingToken, setPendingToken] = useState<string | null>(null)
  const [code, setCode] = useState("")

  useEffect(() => {
    const codeParam = params.get("code")
    const stateParam = params.get("state")
    if (!codeParam || !stateParam) {
      setError("Missing code or state")
      return
    }
    googleCallback.mutateAsync({ code: codeParam, state: stateParam })
      .then((data: any) => {
        if (data.requires2fa && data.tempToken) {
          setPendingToken(data.tempToken)
        } else if (data.token) {
          const { token, ...adminData } = data
          writeSession(adminData as any, token)
          window.location.href = "/admin"
        }
      })
      .catch((e: any) => setError(errorMessage(tc, e)))
  }, [])

  async function handleVerify() {
    if (!pendingToken || !code) return
    try {
      const data: any = await challenge.mutateAsync({ tempToken: pendingToken, code })
      const { token, ...adminData } = data
      writeSession(adminData as any, token)
      window.location.href = "/admin"
    } catch (e: any) {
      setError(errorMessage(tc, e))
    }
  }

  if (error) {
    return (
      <main className="min-h-screen flex items-center justify-center p-6">
        <div className="max-w-sm text-center space-y-4">
          <p className="text-destructive">{error}</p>
          <Button onClick={() => navigate("/login")}>{tc("back")}</Button>
        </div>
      </main>
    )
  }

  if (pendingToken) {
    return (
      <main className="min-h-screen flex items-center justify-center p-6">
        <div className="w-full max-w-sm space-y-4">
          <h1 className="text-xl font-bold">Two-factor authentication</h1>
          <div className="space-y-2">
            <Label htmlFor="code">Code</Label>
            <Input id="code" value={code} onChange={(e) => setCode(e.target.value)} placeholder="123456" />
          </div>
          <Button className="w-full" onClick={handleVerify} disabled={challenge.isPending}>
            {challenge.isPending ? "Verifying…" : "Verify"}
          </Button>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen flex items-center justify-center p-6">
      <p className="text-muted-foreground">{tc("loading")}</p>
    </main>
  )
}
