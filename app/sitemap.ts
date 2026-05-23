import type { MetadataRoute } from 'next';
import { source } from '@/lib/source';
import { docsPathPrefix, siteUrl } from '@/lib/shared';

export const revalidate = false;

export default function sitemap(): MetadataRoute.Sitemap {
  const entries: MetadataRoute.Sitemap = source.getPages().map((page) => {
    const path = page.url === '/' ? '' : page.url;
    return {
      url: `${siteUrl}${docsPathPrefix}${path}`,
      lastModified: page.data.lastModified ?? undefined,
    };
  });

  entries.sort((a, b) => a.url.localeCompare(b.url));
  return entries;
}
