'use client'

import { useAuth } from '@/context/auth-context'
import Link from 'next/link'
import { useActionState, useId, useTransition } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'

const loginSchema = z.object({
  email: z.string().email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
})

type FormState = {
  errors?: Record<string, string>
}

export default function LoginPage() {
  const { login } = useAuth()
  const emailId = useId()
  const passwordId = useId()
  const [isPending, startTransition] = useTransition()

  const [state, formAction] = useActionState<FormState, FormData>(
    async (_prev, formData) => {
      const raw = {
        email: formData.get('email') as string,
        password: formData.get('password') as string,
      }
      const result = loginSchema.safeParse(raw)
      if (!result.success) {
        return { errors: result.error.flatten().fieldErrors as Record<string, string> }
      }
      try {
        await login(result.data)
        toast.success('Welcome back!')
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Login failed')
      }
      return {}
    },
    {},
  )

  return (
    <div className="flex min-h-screen items-center justify-center bg-transparent px-4">
      <div className="w-full max-w-4xl overflow-hidden rounded-2xl border border-white/10 bg-[#1a1a1a] shadow-2xl">
        <div className="grid grid-cols-1 md:grid-cols-2">

          {/* Left — Form */}
          <div className="flex flex-col justify-center px-8 py-12 md:px-12">
            {/* Logo */}
            <div className="mb-8 flex items-center gap-2">
              <img src="/logo.png" alt="Hermes" width={32} height={32} className="object-contain" />
              <span className="text-lg font-semibold text-white">Hermes</span>
            </div>

            <h1 className="text-2xl font-bold text-white">Sign in to your account</h1>
            <p className="mt-1 text-sm text-zinc-400">Enter your credentials to access your relays</p>

            <form action={(fd) => startTransition(() => formAction(fd))} className="mt-8 space-y-5">
              {/* Email */}
              <div>
                <label htmlFor={emailId} className="mb-1.5 block text-sm font-medium text-zinc-300">
                  Email <span className="text-orange-500">*</span>
                </label>
                <input
                  id={emailId}
                  name="email"
                  type="email"
                  autoComplete="email"
                  placeholder="your@email.com"
                  className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
                />
                {state.errors?.email && (
                  <p className="mt-1 text-xs text-red-400">{state.errors.email}</p>
                )}
              </div>

              {/* Password */}
              <div>
                <label htmlFor={passwordId} className="mb-1.5 block text-sm font-medium text-zinc-300">
                  Password <span className="text-orange-500">*</span>
                </label>
                <input
                  id={passwordId}
                  name="password"
                  type="password"
                  autoComplete="current-password"
                  placeholder="Enter your password"
                  className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
                />
                {state.errors?.password && (
                  <p className="mt-1 text-xs text-red-400">{state.errors.password}</p>
                )}
              </div>

              {/* Submit */}
              <button
                type="submit"
                disabled={isPending}
                className="w-full rounded-lg bg-orange-500 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-orange-600 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isPending ? 'Signing in…' : 'Continue'}
              </button>
            </form>

            <p className="mt-6 text-center text-sm text-zinc-500">
              Don&apos;t have an account?{' '}
              <Link href="/register" className="font-medium text-orange-500 hover:text-orange-400">
                Sign up
              </Link>
            </p>
          </div>

          {/* Right — Branding */}
          <div className="hidden flex-col justify-between bg-[#141414] p-12 md:flex">
            <div>
              <h2 className="text-2xl font-bold text-white">Welcome back to Hermes</h2>
              <p className="mt-4 text-sm leading-relaxed text-zinc-400">
                Pick up where you left off. Your relays, integrations, and automations are waiting for you.
              </p>
              <p className="mt-4 text-sm leading-relaxed text-zinc-400">
                Hermes is a webhook relay platform — connect incoming events to Discord, Slack, email, or any HTTP endpoint automatically.
              </p>
              <p className="mt-4 text-sm leading-relaxed text-zinc-400">
                Manage secrets, monitor execution logs, and build powerful automation pipelines in minutes.
              </p>
            </div>

            {/* Decorative illustration */}
            <div className="mt-12 flex items-center justify-center">
              <svg width="180" height="140" viewBox="0 0 180 140" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                {/* Relay node illustration */}
                <rect x="10" y="50" width="50" height="40" rx="8" stroke="#f97316" strokeWidth="1.5" fill="none"/>
                <rect x="60" y="20" width="60" height="100" rx="8" stroke="#f97316" strokeWidth="1.5" fill="none" opacity="0.6"/>
                <rect x="120" y="50" width="50" height="40" rx="8" stroke="#f97316" strokeWidth="1.5" fill="none"/>
                <line x1="60" y1="70" x2="35" y2="70" stroke="#f97316" strokeWidth="1.5" strokeDasharray="4 3"/>
                <line x1="120" y1="70" x2="145" y2="70" stroke="#f97316" strokeWidth="1.5" strokeDasharray="4 3"/>
                <circle cx="90" cy="70" r="12" stroke="#f97316" strokeWidth="1.5" fill="none"/>
                <path d="M85 70 L89 74 L96 66" stroke="#f97316" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
          </div>

        </div>
      </div>
    </div>
  )
}
