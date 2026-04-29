'use client'

import { useCreateSecret, useDeleteSecret, useSecrets } from '@/lib/queries'
import { useId, useActionState, useTransition, useState } from 'react'
import { toast } from 'sonner'

function CreateSecretForm() {
  const nameId = useId()
  const valueId = useId()
  const createMutation = useCreateSecret()
  const [, startTransition] = useTransition()

  const [state, formAction] = useActionState(
    async (_prev: string | null, formData: FormData) => {
      const name = (formData.get('name') as string).trim()
      const value = (formData.get('value') as string).trim()

      if (!name || !value) {
        toast.error('Name and value are required')
        return null
      }

      return new Promise<string | null>((resolve) => {
        startTransition(async () => {
          try {
            await createMutation.mutateAsync({
              name: name.toUpperCase().replace(/\s+/g, '_'),
              value,
            })
            toast.success('Secret saved')
            resolve('ok')
          } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to save secret')
            resolve(null)
          }
        })
      })
    },
    null,
  )

  return (
    <form action={formAction} className="rounded-xl border border-white/10 bg-[#1a1a1a] p-5 space-y-4">
      <h2 className="text-sm font-medium text-zinc-300">Add secret</h2>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label htmlFor={nameId} className="mb-1.5 block text-xs font-medium text-zinc-400">
            Name
          </label>
          <input
            id={nameId}
            name="name"
            type="text"
            key={state}
            defaultValue=""
            placeholder="WEBHOOK_URL"
            className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
          />
        </div>
        <div>
          <label htmlFor={valueId} className="mb-1.5 block text-xs font-medium text-zinc-400">
            Value
          </label>
          <input
            id={valueId}
            name="value"
            type="password"
            key={state}
            defaultValue=""
            placeholder="••••••••••••"
            className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
          />
        </div>
      </div>
      <button
        type="submit"
        disabled={createMutation.isPending}
        className="rounded-lg bg-orange-500 px-4 py-2 text-xs font-semibold text-white transition-colors hover:bg-orange-600 disabled:opacity-50"
      >
        {createMutation.isPending ? 'Saving…' : 'Save secret'}
      </button>
    </form>
  )
}

export default function SecretsPage() {
  const { data: secrets, isLoading } = useSecrets()
  const deleteMutation = useDeleteSecret()
  const [confirmId, setConfirmId] = useState<string | null>(null)

  const handleDelete = async (id: string) => {
    try {
      await deleteMutation.mutateAsync(id)
      toast.success('Secret deleted')
      setConfirmId(null)
    } catch {
      toast.error('Failed to delete secret')
    }
  }

  return (
    <div className="mx-auto w-full max-w-5xl p-8 space-y-8">
      <div>
        <h1 className="text-xl font-bold text-white">Secrets</h1>
        <p className="mt-0.5 text-sm text-zinc-500">Store sensitive values and reference them in relay actions</p>
      </div>

      <CreateSecretForm />

      {/* Secrets list */}
      <div>
        <h2 className="mb-3 text-sm font-medium text-zinc-300">Saved secrets</h2>
        {isLoading ? (
          <div className="space-y-2">
            {[...Array(3)].map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton
              <div key={i} className="h-12 rounded-lg bg-white/5 animate-pulse" />
            ))}
          </div>
        ) : !secrets || secrets.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-[#1a1a1a] px-4 py-8 text-center">
            <p className="text-sm text-zinc-500">No secrets yet. Add one above to use in relay actions.</p>
          </div>
        ) : (
          <div className="rounded-xl border border-white/10 bg-[#1a1a1a] divide-y divide-white/5 overflow-hidden">
            {secrets.map((secret) => (
              <div key={secret.id} className="flex items-center justify-between px-4 py-3">
                <div>
                  <p className="text-sm font-medium text-white font-mono">{secret.name}</p>
                  <p className="text-xs text-zinc-500 mt-0.5">
                    Added {new Date(secret.created_at).toLocaleDateString()}
                  </p>
                </div>
                {confirmId === secret.id ? (
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => handleDelete(secret.id)}
                      disabled={deleteMutation.isPending}
                      className="text-xs text-red-400 hover:text-red-300 transition-colors"
                    >
                      Confirm delete
                    </button>
                    <button
                      type="button"
                      onClick={() => setConfirmId(null)}
                      className="text-xs text-zinc-500 hover:text-zinc-400"
                    >
                      Cancel
                    </button>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={() => setConfirmId(secret.id)}
                    className="text-xs text-zinc-500 hover:text-red-400 transition-colors"
                  >
                    Delete
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <p className="text-xs text-zinc-600 leading-relaxed">
        Secret values are encrypted at rest and never returned by the API. Reference a secret in a relay action using its name (e.g. <code className="text-zinc-400">WEBHOOK_URL</code>).
      </p>
    </div>
  )
}
