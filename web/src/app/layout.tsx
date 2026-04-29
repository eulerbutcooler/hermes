import { AuthProvider } from '@/context/auth-context'
import { QueryProvider } from '@/context/query-provider'
import type { Metadata } from 'next'
import { Montserrat } from 'next/font/google'
import { Toaster } from 'sonner'
import './globals.css'

const montserrat = Montserrat({
  variable: '--font-montserrat',
  subsets: ['latin'],
})

export const metadata: Metadata = {
  title: 'Hermes',
  description: 'Webhook relay & automation platform',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${montserrat.variable} antialiased relative min-h-screen overflow-x-hidden`}>
        {/* Global Watermark Logo */}
        <div className="pointer-events-none fixed -bottom-64 -right-64 z-0 opacity-[0.06]" style={{ mixBlendMode: 'screen' }}>
          <img src="/logo.png" alt="" className="h-[1200px] w-[1200px] object-contain" />
        </div>

        <div className="relative z-10 flex min-h-screen flex-col">
          <QueryProvider>
            <AuthProvider>
              {children}
              <Toaster theme="dark" position="top-right" />
            </AuthProvider>
          </QueryProvider>
        </div>
      </body>
    </html>
  )
}
