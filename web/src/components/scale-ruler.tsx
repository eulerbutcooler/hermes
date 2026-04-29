export function ScaleRuler({ side }: { side: 'left' | 'right' }) {
  const ticks = [0, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550, 600, 650, 700]
  return (
    <div className={`fixed top-0 ${side === 'left' ? 'left-8 border-r' : 'right-8 border-l'} border-white/10 z-50 hidden h-screen w-8 select-none lg:block pointer-events-none`}>
      {/* Orange square at the top */}
      <div className={`absolute top-8 h-1.5 w-1.5 bg-orange-500 ${side === 'left' ? '-right-[3px]' : '-left-[3px]'}`} />
      
      {ticks.map((t) => (
        <div key={t} className={`absolute flex items-center gap-2 ${side === 'left' ? 'right-0 flex-row' : 'left-0 flex-row-reverse'}`} style={{ top: `${(t / 700) * 90 + 10}%` }}>
          <span className="text-[10px] text-zinc-600 font-mono -rotate-90 origin-center tracking-widest">{t}</span>
          <div className="w-1.5 border-t border-zinc-700" />
        </div>
      ))}
    </div>
  )
}
