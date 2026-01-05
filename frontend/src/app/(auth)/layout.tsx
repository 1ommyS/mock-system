export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_var(--page-spot),_transparent_55%)]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,_rgba(20,184,166,0.18),_transparent_45%),_radial-gradient(circle_at_80%_10%,_rgba(250,204,21,0.2),_transparent_35%)]" />
      <div className="pointer-events-none absolute -left-24 top-1/2 h-72 w-72 -translate-y-1/2 rounded-full bg-brand-500/15 blur-[110px]" />
      <div className="pointer-events-none absolute bottom-0 right-0 h-80 w-80 translate-x-1/3 translate-y-1/3 rounded-full bg-amber-300/20 blur-[120px] dark:bg-emerald-400/20" />

      <main className="relative flex min-h-screen items-center justify-center px-6 py-12">
        {children}
      </main>
    </div>
  );
}
