import Link from 'next/link';
import {
  ArrowRight,
  Terminal,
  Cpu,
  HardDrive,
  Layers,
  ShieldCheck,
  Zap,
  BookOpen,
  ExternalLink,
  Code2,
  CheckCircle2,
} from 'lucide-react';

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col bg-neutral-950 text-neutral-100">
      {/* Top Navigation Bar */}
      <header className="sticky top-0 z-50 border-b border-neutral-800 bg-neutral-950/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
          <Link href="/" className="flex items-center gap-2.5 font-medium tracking-tight text-white">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500 font-mono text-sm font-medium text-black">
              TD
            </span>
            <span className="text-base sm:text-lg">tadev</span>
            <span className="rounded bg-neutral-800 px-1.5 py-0.5 text-[10px] font-normal text-neutral-400">
              v1.0.0
            </span>
          </Link>

          <nav className="flex items-center gap-3 sm:gap-6">
            <Link
              href="/docs"
              className="text-xs font-normal text-neutral-300 transition hover:text-white sm:text-sm"
            >
              Docs
            </Link>
            <Link
              href="/docs/java-profiles"
              className="hidden text-xs font-normal text-neutral-300 transition hover:text-white sm:inline sm:text-sm"
            >
              Java Profiles
            </Link>
            <Link
              href="/docs/quickstart"
              className="hidden text-xs font-normal text-neutral-300 transition hover:text-white md:inline sm:text-sm"
            >
              Quickstart
            </Link>
            <a
              href="https://github.com/rugved-danej/termux-android-development"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 rounded-lg border border-neutral-800 bg-neutral-900 px-3 py-1.5 text-xs font-normal text-neutral-200 transition hover:border-neutral-700 hover:text-white sm:text-sm"
            >
              <ExternalLink className="h-3.5 w-3.5" />
              <span>GitHub</span>
            </a>
          </nav>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 flex-col items-center justify-center px-4 py-12 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto w-full max-w-4xl text-center">
          {/* Badge */}
          <div className="mb-6 inline-flex max-w-full flex-wrap items-center justify-center gap-2 rounded-full border border-neutral-800 bg-neutral-900/70 px-4 py-1.5 text-xs font-normal text-neutral-300">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
            <span>Production Release v1.0.0 &bull; Official Termux OpenJDK Support</span>
          </div>

          {/* Heading */}
          <h1 className="text-3xl font-medium tracking-tight text-white sm:text-5xl md:text-6xl">
            Termux Android Development
          </h1>

          {/* Subheading */}
          <p className="mx-auto mt-5 max-w-2xl text-sm leading-6 text-neutral-300 sm:text-base sm:leading-7 md:text-lg">
            Automated, production-grade Android SDK and OpenJDK build environment for Termux on Android devices.
            Compile real Android APKs and AABs directly on mobile hardware without root or PRoot.
          </p>

          {/* Action Buttons */}
          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4">
            <Link
              href="/docs"
              className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 px-6 py-3.5 text-sm font-medium text-black shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 sm:w-auto"
            >
              <BookOpen className="h-4 w-4" />
              <span>Explore Documentation</span>
              <ArrowRight className="h-4 w-4" />
            </Link>
            <a
              href="https://github.com/rugved-danej/termux-android-development"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-neutral-800 bg-neutral-900 px-6 py-3.5 text-sm font-medium text-white transition hover:border-neutral-700 hover:bg-neutral-800 sm:w-auto"
            >
              <Code2 className="h-4 w-4 text-emerald-400" />
              <span>GitHub Repository</span>
            </a>
          </div>

          {/* Quick Install Command Box */}
          <div className="mt-10 w-full overflow-hidden rounded-xl border border-neutral-800 bg-neutral-950/90 text-left shadow-2xl backdrop-blur sm:mt-12">
            <div className="flex items-center justify-between border-b border-neutral-800/80 px-4 py-2.5 text-xs text-neutral-400">
              <div className="flex items-center gap-2">
                <Terminal className="h-3.5 w-3.5 text-emerald-400" />
                <span className="font-normal">One-Line Automated Installer</span>
              </div>
              <span className="font-mono text-[11px] text-neutral-500">Termux Shell</span>
            </div>
            <div className="overflow-x-auto p-3 sm:p-4">
              <pre className="font-mono text-xs text-emerald-300 sm:text-sm">
                <code>bash &lt;(curl -sL https://raw.githubusercontent.com/rugved-danej/termux-android-development/main/install.sh)</code>
              </pre>
            </div>
          </div>

          {/* Feature Highlights Grid */}
          <div className="mt-14 grid grid-cols-1 gap-4 text-left sm:grid-cols-2 sm:gap-6 lg:grid-cols-3 sm:mt-16">
            <div className="rounded-xl border border-neutral-800/80 bg-neutral-900/40 p-5 sm:p-6 transition hover:border-neutral-700">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
                <Cpu className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-sm font-medium text-white sm:text-base">Native ARM64 AAPT2 Override</h3>
              <p className="mt-1.5 text-xs text-neutral-400 sm:text-sm">
                Overrides x86_64 maven binaries with Termux native ARM64 AAPT2 via global Gradle properties.
              </p>
            </div>

            <div className="rounded-xl border border-neutral-800/80 bg-neutral-900/40 p-5 sm:p-6 transition hover:border-neutral-700">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
                <Layers className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-sm font-medium text-white sm:text-base">Multi-Profile Java Engine</h3>
              <p className="mt-1.5 text-xs text-neutral-400 sm:text-sm">
                Official OpenJDK 21 LTS (recommended), 25, and 17. Zero third-party repositories or unstable packages.
              </p>
            </div>

            <div className="rounded-xl border border-neutral-800/80 bg-neutral-900/40 p-5 sm:p-6 transition hover:border-neutral-700">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
                <Zap className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-sm font-medium text-white sm:text-base">Dynamic RAM Auto-Tuning</h3>
              <p className="mt-1.5 text-xs text-neutral-400 sm:text-sm">
                Reads device /proc/meminfo to configure optimal Gradle daemon heap sizes and prevent Android LMK termination.
              </p>
            </div>

            <div className="rounded-xl border border-neutral-800/80 bg-neutral-900/40 p-5 sm:p-6 transition hover:border-neutral-700">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
                <HardDrive className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-base font-medium text-white">Storage Optimization</h3>
              <p className="mt-1.5 text-xs text-neutral-400 sm:text-sm">
                Installs exclusively in private Termux app storage, bypassing Android /sdcard exec restrictions.
              </p>
            </div>

            <div className="rounded-xl border border-neutral-800/80 bg-neutral-900/40 p-5 sm:p-6 transition hover:border-neutral-700">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
                <ShieldCheck className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-base font-medium text-white">Zero Termux:API Dependency</h3>
              <p className="mt-1.5 text-xs text-neutral-400 sm:text-sm">
                Native browser opening via Activity Manager intents without requiring external companion APKs.
              </p>
            </div>

            <div className="rounded-xl border border-neutral-800/80 bg-neutral-900/40 p-5 sm:p-6 transition hover:border-neutral-700">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
                <Terminal className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-base font-medium text-white">Interactive TUI Dashboard</h3>
              <p className="mt-1.5 text-xs text-neutral-400 sm:text-sm">
                Full keyboard-driven terminal user interface for seamless environment management and diagnostics.
              </p>
            </div>
          </div>

          {/* Java Profiles Section */}
          <div className="mt-16 rounded-2xl border border-neutral-800 bg-neutral-900/30 p-6 text-left sm:p-8">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 className="text-xl font-medium text-white sm:text-2xl">Official Termux Java Profiles</h2>
                <p className="mt-1 text-xs text-neutral-400 sm:text-sm">
                  Exclusively official Termux repository packages. Zero tur-repo dependency.
                </p>
              </div>
              <Link
                href="/docs/java-profiles"
                className="inline-flex items-center gap-1.5 text-xs font-normal text-emerald-400 hover:text-emerald-300 sm:text-sm"
              >
                <span>View Full Comparison</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </Link>
            </div>

            <div className="mt-6 grid grid-cols-1 gap-4 md:grid-cols-3">
              {/* Profile 1: OpenJDK 21 LTS (Recommended) */}
              <div className="relative rounded-xl border border-emerald-500/50 bg-emerald-950/10 p-5">
                <div className="absolute -top-2.5 right-4 rounded-full bg-emerald-500 px-2 py-0.5 text-[10px] font-normal text-black uppercase tracking-wider">
                  Recommended
                </div>
                <div className="text-sm font-medium text-white">openjdk-21</div>
                <div className="mt-1 text-xs text-emerald-400">Java 21 LTS</div>
                <p className="mt-2 text-xs text-neutral-400">
                  Default and golden standard for AGP 8.x and Gradle 8.x. Best balance of speed and stability.
                </p>
                <div className="mt-4 flex items-center gap-2 text-[11px] text-neutral-300">
                  <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
                  <span>AGP 8.0 - 8.11+ Support</span>
                </div>
              </div>

              {/* Profile 2: OpenJDK 25 */}
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/50 p-5">
                <div className="text-sm font-medium text-white">openjdk-25</div>
                <div className="mt-1 text-xs text-neutral-400">Java 25 (Latest STS)</div>
                <p className="mt-2 text-xs text-neutral-400">
                  Cutting-edge Java runtime with latest language innovations and experimental features.
                </p>
                <div className="mt-4 flex items-center gap-2 text-[11px] text-neutral-300">
                  <CheckCircle2 className="h-3.5 w-3.5 text-neutral-500" />
                  <span>Latest Features</span>
                </div>
              </div>

              {/* Profile 3: OpenJDK 17 LTS */}
              <div className="rounded-xl border border-neutral-800 bg-neutral-900/50 p-5">
                <div className="text-sm font-medium text-white">openjdk-17</div>
                <div className="mt-1 text-xs text-neutral-400">Java 17 LTS (Legacy)</div>
                <p className="mt-2 text-xs text-neutral-400">
                  Maximum compatibility for legacy projects on AGP 7.x and older Gradle distributions.
                </p>
                <div className="mt-4 flex items-center gap-2 text-[11px] text-neutral-300">
                  <CheckCircle2 className="h-3.5 w-3.5 text-neutral-500" />
                  <span>Legacy AGP Compatibility</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Responsive Footer */}
      <footer className="border-t border-neutral-800 bg-neutral-950 py-8 text-neutral-500">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 px-4 sm:flex-row sm:px-6 lg:px-8">
          <div className="flex items-center gap-2 text-xs">
            <span className="font-mono font-medium text-neutral-400">tadev</span>
            <span>&bull;</span>
            <span>Production Termux Android Development Toolchain</span>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-4 text-xs">
            <Link href="/docs" className="hover:text-neutral-300">
              Documentation
            </Link>
            <Link href="/docs/quickstart" className="hover:text-neutral-300">
              Quickstart
            </Link>
            <Link href="/docs/java-profiles" className="hover:text-neutral-300">
              Java Profiles
            </Link>
            <Link href="/docs/architecture" className="hover:text-neutral-300">
              Architecture
            </Link>
            <a
              href="https://github.com/rugved-danej/termux-android-development"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-neutral-300"
            >
              GitHub
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
