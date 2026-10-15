import { docs } from '@/.source';
import { loader } from 'fumadocs-core/source';

const rawSource = docs.toFumadocsSource() as any;

export const source = loader({
  baseUrl: '/docs',
  source: {
    ...rawSource,
    files: typeof rawSource.files === 'function' ? rawSource.files() : rawSource.files,
  },
});
