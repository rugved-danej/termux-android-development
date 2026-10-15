import { RootProvider } from 'fumadocs-ui/provider/next';
import './global.css';
import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';

export const metadata: Metadata = {
  metadataBase: new URL('https://tadev.vercel.app'),
  title: {
    template: '%s | Termux Android Development (tadev)',
    default: 'Termux Android Development (tadev) Documentation',
  },
  description: 'Production-grade Android SDK and OpenJDK build environment for Termux on Android.',
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="flex min-h-screen flex-col bg-neutral-950 text-neutral-100 antialiased selection:bg-emerald-500 selection:text-black">
        <RootProvider>
          {children}
        </RootProvider>
      </body>
    </html>
  );
}
