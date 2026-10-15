import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import { source } from '@/lib/source';
import type { ReactNode } from 'react';

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <DocsLayout
      tree={source.getPageTree()}
      nav={{
        title: 'tadev Documentation',
        url: '/docs',
      }}
      links={[
        {
          text: 'GitHub Repository',
          url: 'https://github.com/rugved-danej/termux-android-development',
          active: 'nested-url',
          external: true,
        },
      ]}
    >
      {children}
    </DocsLayout>
  );
}
