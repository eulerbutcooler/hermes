'use client'

import Link from 'next/link'
import { useState } from 'react'

import { ScaleRuler } from '@/components/scale-ruler'
/* ─── Section divider with dots (from the design) ─── */
function SectionDivider() {
  return (
    <div className="flex items-center justify-center py-8">
      <div className="flex items-center gap-0">
        <div className="h-[3px] w-[3px] rounded-full bg-orange-500" />
        <div className="h-[1px] w-40 bg-gradient-to-r from-orange-500/60 to-orange-500/20" />
        <div className="h-[3px] w-[3px] rounded-full bg-orange-500" />
        <div className="mx-5 flex items-center justify-center">
          <img src="/logo.png" alt="" className="h-10 w-10 object-contain drop-shadow-md" />
        </div>
        <div className="h-[3px] w-[3px] rounded-full bg-orange-500" />
        <div className="h-[1px] w-40 bg-gradient-to-l from-orange-500/60 to-orange-500/20" />
        <div className="h-[3px] w-[3px] rounded-full bg-orange-500" />
      </div>
    </div>
  )
}

/* ─── FAQ Item ─── */
function FAQItem({ question, answer }: { question: string; answer: string }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="border-b border-white/10">
      <button type="button" onClick={() => setOpen(!open)} className="flex w-full items-center justify-between py-5 text-left text-sm font-medium text-white transition-colors hover:text-orange-400">
        {question}
        <svg className={`h-4 w-4 shrink-0 text-zinc-500 transition-transform ${open ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="m19 9-7 7-7-7" /></svg>
      </button>
      {open && <p className="pb-5 text-sm leading-relaxed text-zinc-400">{answer}</p>}
    </div>
  )
}

export default function LandingPage() {
  return (
    <div className="relative min-h-screen overflow-hidden bg-[#111111] text-white">
      <ScaleRuler side="left" />
      <ScaleRuler side="right" />

      {/* ═══════ NAVBAR ═══════ */}
      <nav className="sticky top-4 z-40 mx-auto mt-4 flex max-w-3xl items-center justify-between rounded-full border border-white/10 bg-[#1a1a1a]/80 px-5 py-2.5 backdrop-blur-xl">
        <div className="flex items-center gap-2">
          <img src="/logo.png" alt="Hermes" className="h-10 w-10 object-contain drop-shadow-md" />
        </div>
        <div className="hidden items-center gap-6 text-sm text-zinc-400 sm:flex">
          <a href="#features" className="transition-colors hover:text-white">Features</a>
          <a href="#demo" className="transition-colors hover:text-white">Demo</a>
          <a href="#pricing" className="transition-colors hover:text-white">Pricing</a>
          <a href="#faq" className="transition-colors hover:text-white">FAQ</a>
        </div>
        <Link href="/login" className="rounded-full bg-orange-500 px-4 py-1.5 text-xs font-semibold text-white transition-all hover:bg-orange-600">
          Login →
        </Link>
      </nav>

      {/* ═══════ HERO ═══════ */}
      <section className="mx-auto max-w-6xl px-6 pb-20 pt-24 lg:px-20">
        <div className="grid items-center gap-12 lg:grid-cols-2">
          <div>
            <h1 className="text-4xl font-bold leading-tight tracking-tight text-white sm:text-5xl lg:text-[3.2rem]" style={{ fontFamily: 'Georgia, serif' }}>
              Your workflows aren&apos;t linear.<br />
              Your engine shouldn&apos;t be either.
            </h1>
            <p className="mt-6 max-w-lg text-base leading-relaxed text-zinc-400">
              A DAG-based automation engine with webhook triggers, cron scheduling, and multi-service execution built for production systems.
            </p>
            <div className="mt-8 flex flex-wrap items-center gap-4">
              <Link href="/register" className="inline-flex items-center gap-2 rounded-full bg-orange-500 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-orange-500/25 transition-all hover:bg-orange-600 hover:shadow-orange-500/40">
                Get Started <span>›</span>
              </Link>
              <a href="https://github.com" target="_blank" rel="noreferrer" className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-6 py-3 text-sm font-medium text-white transition-all hover:border-white/25 hover:bg-white/10">
                Star Github
                <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              </a>
            </div>
          </div>
          <div className="relative hidden lg:block">
            <img src="/logo.png" alt="Hermes" className="absolute top-1/2 -translate-y-1/2 -right-48 w-[60rem] max-w-none opacity-20 pointer-events-none" style={{ mixBlendMode: 'screen' }} />
          </div>
        </div>
      </section>

      <SectionDivider />

      {/* ═══════ FEATURES ═══════ */}
      <section id="features" className="mx-auto max-w-6xl px-6 py-20 lg:px-20">
        <div className="mb-4 flex justify-center">
          <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-4 py-1.5 text-xs font-medium text-zinc-300">
            ✦ Features
          </span>
        </div>
        <p className="mb-14 text-center text-lg italic text-zinc-400" style={{ fontFamily: 'Georgia, serif' }}>
          Everything you need to build powerful automations
        </p>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {/* Card 1 */}
          <div className="row-span-2 rounded-2xl border border-white/8 bg-[#161616] p-8">
            <div className="mb-6 flex h-48 items-center justify-center rounded-xl bg-white/[0.02]">
              <svg className="h-24 w-24 text-zinc-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={0.5}><path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" /></svg>
            </div>
            <h3 className="mb-2 text-xl font-bold text-white">Connect everything you already use</h3>
            <p className="text-sm leading-relaxed text-zinc-400">Bring Discord, Slack, Email, HTTP APIs, and internal tools into one unified workflow.</p>
          </div>
          {/* Card 2 */}
          <div className="rounded-2xl border border-white/8 bg-[#161616] p-8">
            <div className="mb-4 flex items-center gap-2 rounded-lg border border-orange-500/20 bg-orange-500/5 px-3 py-2 text-xs text-orange-400">
              <div className="h-3 w-3 animate-spin rounded-full border border-orange-400 border-t-transparent" />
              2 Jobs running
            </div>
            <h3 className="mb-2 text-lg font-bold text-white">Scheduling &amp; Workflows</h3>
            <p className="text-sm text-zinc-400">Schedule workflows with cron, trigger via webhooks, events, or manually.</p>
          </div>
          {/* Card 3 */}
          <div className="rounded-2xl border border-white/8 bg-[#161616] p-8">
            <h3 className="mb-2 text-lg font-bold text-white">Start with proven templates</h3>
            <p className="text-sm text-zinc-400">Launch common workflows in one click, customize later.</p>
          </div>
          {/* Card 4 */}
          <div className="rounded-2xl border border-white/8 bg-[#161616] p-8">
            <div className="mb-4 grid grid-cols-2 gap-2">
              {['AI', '⚡', '🔗', '📊'].map((icon) => (
                <div key={icon} className="flex h-12 items-center justify-center rounded-lg border border-white/5 bg-white/[0.02] text-lg">{icon}</div>
              ))}
            </div>
            <h3 className="mb-2 text-lg font-bold text-white">Extensible integrations</h3>
            <p className="text-sm text-zinc-400">Build custom action types and extend the platform with your own services.</p>
          </div>
          {/* Card 5 */}
          <div className="rounded-2xl border border-white/8 bg-[#161616] p-8">
            <h3 className="mb-2 text-lg font-bold text-white">DAG-based execution</h3>
            <p className="text-sm text-zinc-400">Model complex branching logic with conditions, parallel actions, and dependency chains.</p>
          </div>
        </div>
      </section>

      <SectionDivider />

      {/* ═══════ DEMO ═══════ */}
      <section id="demo" className="mx-auto max-w-6xl px-6 py-20 lg:px-20">
        <div className="mb-4 flex justify-center">
          <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-4 py-1.5 text-xs font-medium text-zinc-300">
            ✧ Playground
          </span>
        </div>
        <p className="mb-14 text-center text-lg italic text-zinc-400" style={{ fontFamily: 'Georgia, serif' }}>
          See how it all comes together
        </p>
        <div className="overflow-hidden rounded-2xl border border-white/10 bg-[#0d0d0d]">
          <div className="flex items-center gap-2 border-b border-white/10 px-4 py-3">
            <div className="h-3 w-3 rounded-full bg-red-500/80" />
            <div className="h-3 w-3 rounded-full bg-yellow-500/80" />
            <div className="h-3 w-3 rounded-full bg-green-500/80" />
            <span className="ml-auto text-xs text-zinc-500">hermes.app</span>
          </div>
          <div className="flex min-h-[350px] items-center justify-center p-12">
            <div className="flex flex-wrap items-center justify-center gap-6">
              {[
                { name: 'Webhook Trigger', sub: 'POST /hook/abc', icon: '⚡' },
                { name: 'Validate Payload', sub: 'Condition check', icon: '🔍' },
                { name: 'Send to Slack', sub: 'Channel: #alerts', icon: '💬' },
                { name: 'Log to DB', sub: 'Debug log', icon: '📝' },
              ].map((node, i) => (
                <div key={node.name} className="flex items-center gap-4">
                  <div className="rounded-xl border border-dashed border-white/15 bg-[#1a1a1a] p-4 text-center">
                    <div className="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-lg">{node.icon}</div>
                    <p className="text-xs font-medium text-white">{node.name}</p>
                    <p className="text-[10px] text-zinc-500">{node.sub}</p>
                  </div>
                  {i < 3 && <span className="text-zinc-600">→</span>}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <SectionDivider />

      {/* ═══════ PRICING ═══════ */}
      <section id="pricing" className="mx-auto max-w-6xl px-6 py-20 lg:px-20">
        <h2 className="mb-3 text-center text-3xl font-bold italic tracking-tight lg:text-4xl" style={{ fontFamily: 'Georgia, serif' }}>
          Pricing that scales with your business
        </h2>
        <p className="mb-10 text-center text-sm text-zinc-400">Choose the perfect plan for your needs<br />and start optimizing your workflow today</p>

        <div className="mb-10 flex items-center justify-center gap-1 rounded-full border border-white/10 bg-white/5 p-1 mx-auto w-fit">
          <span className="rounded-full px-4 py-1.5 text-xs text-zinc-400">Monthly</span>
          <span className="rounded-full bg-white px-4 py-1.5 text-xs font-medium text-black">Annually</span>
        </div>
        <p className="mb-10 text-center text-xs text-orange-400">Save 25% <span className="text-zinc-500">On Annual Billing</span></p>

        <div className="grid gap-6 sm:grid-cols-3">
          {[
            { name: 'Free', desc: 'For developers trying out Hermes for the first time', price: '$0', features: ['Up to 5 relays', '500 executions per month', 'Basic integrations', 'Community support'], highlighted: false },
            { name: 'Pro', desc: 'Ideal for developers who need more features and support', price: '$15', features: ['Unlimited relays', '10,000 executions per month', 'All premium integrations', 'Priority support'], highlighted: true },
            { name: 'Startup', desc: 'For teams scaling their automation workflows', price: '$38', features: ['50,000 executions per month', 'Team collaboration (up to 5 users)', 'Custom webhooks', 'Dedicated support'], highlighted: false },
          ].map((plan) => (
            <div key={plan.name} className={`rounded-2xl border p-8 ${plan.highlighted ? 'border-orange-500/30 bg-[#1a1a1a]' : 'border-white/8 bg-[#141414]'}`}>
              <h3 className="text-lg font-bold text-white">{plan.name}</h3>
              <p className="mt-1 text-xs text-zinc-500">{plan.desc}</p>
              <p className="mt-6 text-4xl font-bold text-white">{plan.price}</p>
              <p className="text-xs text-zinc-500">Per month</p>
              <Link href="/register" className={`mt-6 block rounded-lg py-2.5 text-center text-sm font-medium transition-all ${plan.highlighted ? 'bg-orange-500 text-white hover:bg-orange-600' : 'border border-white/10 text-zinc-300 hover:border-white/20 hover:text-white'}`}>
                Get Started
              </Link>
              <ul className="mt-6 space-y-3">
                {plan.features.map((f) => (
                  <li key={f} className="flex items-start gap-2 text-sm text-zinc-400">
                    <svg className="mt-0.5 h-4 w-4 shrink-0 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
                    {f}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </section>

      <SectionDivider />

      {/* ═══════ FAQ ═══════ */}
      <section id="faq" className="mx-auto max-w-6xl px-6 py-20 lg:px-20">
        <div className="grid gap-12 lg:grid-cols-[1fr_2fr]">
          <div>
            <p className="text-xs font-medium tracking-widest text-zinc-500">F.A.Q</p>
            <h2 className="mt-2 text-3xl font-bold italic tracking-tight" style={{ fontFamily: 'Georgia, serif' }}>Frequently Asked<br />Questions.</h2>
            <p className="mt-4 text-sm text-zinc-400">Get <strong className="text-white">answers</strong> to <strong className="text-white">commonly</strong> asked questions.</p>
          </div>
          <div>
            <FAQItem question="What is Hermes?" answer="Hermes is a DAG-based automation platform that connects different services and executes automated workflows via webhooks and cron schedules. It supports Discord, Slack, Email, HTTP requests, and more." />
            <FAQItem question="How many relays can I create?" answer="On the free plan you can create up to 5 relays. Pro and Startup plans offer unlimited relays with higher execution limits." />
            <FAQItem question="What is an execution?" answer="An execution is a single run of a relay workflow. Each time a webhook fires or a cron schedule triggers, it creates one execution that processes through your configured action chain." />
            <FAQItem question="Can I integrate with my existing tools?" answer="Yes! Hermes supports Discord webhooks, Slack messages, email via Google and Microsoft OAuth, arbitrary HTTP requests, and debug logging out of the box." />
            <FAQItem question="Do I need coding experience to use Hermes?" answer="No. The visual DAG builder lets you drag and drop components to create workflows. For advanced users, you can configure custom HTTP requests and conditions." />
            <FAQItem question="Can I upgrade or downgrade my plan?" answer="Absolutely. You can change your plan at any time from the dashboard settings. Changes take effect immediately." />
          </div>
        </div>
      </section>

      <SectionDivider />

      {/* ═══════ CTA ═══════ */}
      <section className="mx-auto max-w-6xl px-6 py-16 lg:px-20">
        <div className="overflow-hidden rounded-3xl bg-gradient-to-br from-orange-500 to-orange-600 px-8 py-16 text-center shadow-2xl shadow-orange-500/20">
          <h2 className="text-3xl font-bold text-white lg:text-4xl">Ready to automate your workflow?</h2>
          <p className="mt-3 text-base text-white/80">Start building powerful automations in minutes with Hermes</p>
          <Link href="/register" className="mt-8 inline-flex items-center gap-2 rounded-full bg-white px-8 py-3 text-sm font-semibold text-orange-600 shadow-lg transition-all hover:bg-zinc-100">
            Get Started <span>›</span>
          </Link>
        </div>
      </section>

      {/* ═══════ FOOTER ═══════ */}
      <footer className="mx-auto max-w-6xl px-6 pb-20 pt-16 lg:px-20">
        {/* Footer links grid */}
        <div className="grid grid-cols-2 gap-8 border-t border-white/10 pt-12 sm:grid-cols-3 lg:grid-cols-5">
          {[
            { title: 'Product', links: ['Workflows', 'Integrations', 'Templates', 'Pricing'] },
            { title: 'Resources', links: ['Documentation', 'Guides', 'Blog', 'Tutorials'] },
            { title: 'Developers', links: ['API Reference', 'SDKs', 'Webhooks', 'Examples'] },
            { title: 'Community', links: ['Discord', 'GitHub', 'Support', 'Status'] },
            { title: 'Company', links: ['About', 'Contact', 'Careers'] },
          ].map((col) => (
            <div key={col.title}>
              <h4 className="mb-4 text-sm font-semibold text-white">{col.title}</h4>
              <ul className="space-y-2.5">{col.links.map((l) => <li key={l}><span className="cursor-pointer text-sm text-zinc-500 transition-colors hover:text-white">{l}</span></li>)}</ul>
            </div>
          ))}
        </div>

        {/* Bottom bar */}
        <div className="mt-12 flex flex-col items-center justify-between gap-4 border-t border-white/10 pt-8 sm:flex-row">
          <p className="text-xs text-zinc-500">© 2026 Hermes. All rights reserved.</p>
          <div className="flex gap-4">
            <span className="cursor-pointer text-xs text-zinc-500 hover:text-white">Privacy Policy</span>
            <span className="cursor-pointer text-xs text-zinc-500 hover:text-white">Terms of Service</span>
          </div>
        </div>

        {/* Large brand text at bottom */}
        <div className="mt-16 overflow-hidden">
          <h2 className="text-center text-[8rem] font-black uppercase leading-none tracking-tighter text-white/[0.03] sm:text-[12rem]">
            HERMES
          </h2>
        </div>
      </footer>
    </div>
  )
}
