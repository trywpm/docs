import type { BaseIndex, DocumentRecord } from 'fumadocs-core/search/algolia';

import { getBreadcrumbItems } from 'fumadocs-core/breadcrumb';

import { source } from '@/lib/source';

function toIndex(page: DocumentRecord): BaseIndex[] {
  let id = 0;
  const records: BaseIndex[] = [];
  const seenHeadings = new Set<string>();

  function create(section: string | undefined, sectionId: string | undefined, content: string) {
    return {
      objectID: `${page._id}-${(id++).toString()}`,
      breadcrumbs: page.breadcrumbs,
      title: page.title,
      url: page.url,
      page_id: page._id,
      tag: page.tag,
      section,
      section_id: sectionId,
      content,
    } satisfies BaseIndex;
  }

  if (page.description) {
    records.push(create(undefined, undefined, page.description));
  }

  const { headings, contents } = page.structured;
  for (const paragraph of contents) {
    const heading = paragraph.heading
      ? headings.find((h) => h.id === paragraph.heading)
      : undefined;

    if (heading && !seenHeadings.has(heading.id)) {
      seenHeadings.add(heading.id);
      records.push(create(heading.content, heading.id, heading.content));
    }

    records.push(create(heading?.content, heading?.id, paragraph.content));
  }

  return records;
}

export function exportSearchIndexes(): BaseIndex[] {
  const records: BaseIndex[] = [];

  for (const page of source.getPages()) {
    const breadcrumbs = getBreadcrumbItems(page.url, source.pageTree)
      .map((item) => (typeof item.name === 'string' ? item.name : undefined))
      .filter((name): name is string => name !== undefined);

    records.push(
      ...toIndex({
        _id: page.url,
        structured: page.data.structuredData,
        url: page.url,
        title: page.data.title,
        description: page.data.description,
        breadcrumbs: breadcrumbs.length > 0 ? breadcrumbs : undefined,
      }),
    );
  }

  return records;
}
