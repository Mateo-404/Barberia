import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Shield, KeyRound } from "lucide-react"
import { useTOTPSetup, useTOTPVerify, useTOTPDisable } from "@/api/auth"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { errorMessage } from "@/lib/error-message"

export default function SecurityPage() {
  const { t } = useTranslation("admin")
  const { t: tc } = useTranslation()
  const setup = useTOTPSetup()
  const verify = useTOTPVerify()
  const disable = useTOTPDisable()
  const [secret, setSecret] = useState<string | null>(null)
  const [otpauth, setOtpauth] = useState<string | null>(null)
  const [code, setCode] = useState("")
  const [msg, setMsg] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  async function handleSetup() {
    setErr(null); setMsg(null)
    try {
      const data: any = await setup.mutateAsync()
      setSecret(data.secret)
      setOtpauth(data.otpauthUrl)
    } catch (e: any) {
      setErr(errorMessage(tc, e))
    }
  }

  async function handleVerify() {
    setErr(null); setMsg(null)
    try {
      await verify.mutateAsync(code)
      setMsg("Two-factor authentication enabled.")
      setSecret(null); setOtpauth(null); setCode("")
    } catch (e: any) {
      setErr(errorMessage(tc, e))
    }
  }

  async function handleDisable() {
    setErr(null); setMsg(null)
    try {
      await disable.mutateAsync(code)
      setMsg("Two-factor authentication disabled.")
      setCode("")
    } catch (e: any) {
      setErr(errorMessage(tc, e))
    }
  }

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-6 sm:px-6">
      <header className="mb-8 border-b border-border/60 pb-6">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
          {t("settings.sections.security.title")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("settings.sections.security.description")}
        </p>
      </header>

      <section className="space-y-4 rounded-lg border border-border/60 bg-card/40 p-6">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
            <Shield className="h-5 w-5" />
          </span>
          <div>
            <p className="text-sm font-medium text-foreground">Authenticator app (TOTP)</p>
            <p className="text-xs text-muted-foreground">Use Google Authenticator, Authy or any TOTP app.</p>
          </div>
        </div>

        {!secret && !otpauth && (
          <Button onClick={handleSetup} disabled={setup.isPending}>
            <KeyRound className="mr-2 h-4 w-4" />
            {setup.isPending ? "Generating…" : "Set up 2FA"}
          </Button>
        )}

        {secret && otpauth && (
          <div className="space-y-4">
            <div className="rounded-md bg-muted p-3">
              <p className="text-xs text-muted-foreground mb-1">Secret (enter manually if you cannot scan a QR):</p>
              <code className="block break-all font-mono text-sm text-foreground">{secret}</code>
            </div>
            <p className="text-xs text-muted-foreground break-all">
              otpauth URL: <code className="font-mono">{otpauth}</code>
            </p>
            <div className="space-y-2">
              <Label htmlFor="verify-code">Enter the 6-digit code to confirm</Label>
              <Input
                id="verify-code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="123456"
                className="text-base text-center tracking-widest"
                value={code}
                onChange={(e) => setCode(e.target.value)}
              />
            </div>
            <Button onClick={handleVerify} disabled={verify.isPending || !code}>
              {verify.isPending ? "Verifying…" : "Enable 2FA"}
            </Button>
          </div>
        )}

        <div className="border-t border-border/60 pt-4 space-y-2">
          <p className="text-sm font-medium text-foreground">Disable 2FA</p>
          <div className="flex gap-2">
            <Input
              type="text"
              inputMode="numeric"
              placeholder="123456"
              className="text-base text-center tracking-widest"
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
            <Button variant="destructive" onClick={handleDisable} disabled={disable.isPending || !code}>
              Disable
            </Button>
          </div>
        </div>

        {err && <p role="alert" className="text-sm text-destructive">{err}</p>}
        {msg && <p className="text-sm text-success">{msg}</p>}
      </section>
    </div>
  )
}
