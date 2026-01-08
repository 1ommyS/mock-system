import Link from "next/link";

export default function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_var(--page-spot),_transparent_55%)]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,_rgba(56,189,248,0.12),_transparent_45%),_radial-gradient(circle_at_80%_10%,_rgba(148,163,184,0.14),_transparent_35%)]" />
      <header className="relative z-10 flex items-center justify-between px-8 py-6">
        <Link className="text-lg font-semibold" href="/mocks">
          Mock Storage
        </Link>
        <nav className="flex items-center gap-4 text-sm text-[var(--text-muted)]">
          <Link className="hover:text-[var(--text-primary)]" href="/mocks">
            Моки
          </Link>
          <Link className="hover:text-[var(--text-primary)]" href="/dsl-runner">
            DSL Runner
          </Link>
          <Link className="hover:text-[var(--text-primary)]" href="/profile">
            Профиль
          </Link>
        </nav>
      </header>
      <main className="relative z-10 px-6 pb-12">{children}</main>
    </div>
  );
}
