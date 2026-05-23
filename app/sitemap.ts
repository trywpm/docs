import type { MetadataRoute } from 'next';
import { source } from '@/lib/source';
import { docsPathPrefix, siteUrl } from '@/lib/shared';

export const revalidate = false;

type Page = ReturnType<typeof source.getPages>[number];

function priorityFor(page: Page): number {
  const slugs = page.slugs;
  if (slugs.length === 2 && slugs[0] === 'getting-started' && slugs[1] === 'overview') {
    return 1.0;
  }
  if (slugs.length <= 1) {
    return 0.9;
  }
  if (slugs.length === 2) {
    return 0.8;
  }
  if (slugs.length === 3) {
    return 0.7;
  }
  return 0.6;
}

function changeFrequencyFor(page: Page): MetadataRoute.Sitemap[number]['changeFrequency'] {
  return page.slugs[0] === 'reference' ? 'monthly' : 'weekly';
}

export default function sitemap(): MetadataRoute.Sitemap {
  const entries: MetadataRoute.Sitemap = source.getPages().map((page) => ({
    url: `${siteUrl}${docsPathPrefix}${page.url}`,
    lastModified: page.data.lastModified ?? undefined,
    changeFrequency: changeFrequencyFor(page),
    priority: priorityFor(page),
  }));

  entries.sort((a, b) => a.url.localeCompare(b.url));
  return entries;
}
