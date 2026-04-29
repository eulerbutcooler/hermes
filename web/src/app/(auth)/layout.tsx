import { ScaleRuler } from '@/components/scale-ruler'

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <ScaleRuler side="left" />
      <ScaleRuler side="right" />
      {children}
    </>
  )
}
