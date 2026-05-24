'use client';

import type { SharedProps } from 'fumadocs-ui/components/dialog/search';
import type { BaseIndex } from 'fumadocs-core/search/algolia';
import type { SearchClient } from 'fumadocs-core/search/client';
import type { SortedResult } from 'fumadocs-core/search';

import { liteClient } from 'algoliasearch/lite';
import {
  SearchDialog,
  SearchDialogIcon,
  SearchDialogList,
  SearchDialogInput,
  SearchDialogClose,
  SearchDialogHeader,
  SearchDialogContent,
  SearchDialogOverlay,
} from 'fumadocs-ui/components/dialog/search';
import { useDocsSearch } from 'fumadocs-core/search/client';
import { createContentHighlighter } from 'fumadocs-core/search';

import { algoliaIndexName, algoliaAppId, algoliaSearchKey } from '@/lib/shared';

const client = liteClient(algoliaAppId, algoliaSearchKey);

const searchClient: SearchClient = {
  deps: [algoliaIndexName],
  async search(query) {
    if (query.trim().length === 0) {
      return [];
    }

    const { results } = await client.searchForHits<BaseIndex>({
      requests: [
        {
          type: 'default',
          indexName: algoliaIndexName,
          query,
          distinct: 5,
          hitsPerPage: 10,
        },
      ],
    });

    const highlighter = createContentHighlighter(query);
    const seenUrls = new Set<string>();
    const sorted: SortedResult[] = [];

    for (const hit of results[0].hits) {
      if (!seenUrls.has(hit.url)) {
        seenUrls.add(hit.url);
        sorted.push({
          id: hit.url,
          type: 'page',
          url: hit.url,
          breadcrumbs: hit.breadcrumbs,
          content: highlighter.highlightMarkdown(hit.title),
        });
      }

      const isHeading = hit.section !== undefined && hit.content === hit.section;
      sorted.push({
        id: hit.objectID,
        type: isHeading ? 'heading' : 'text',
        url: hit.section_id ? `${hit.url}#${hit.section_id}` : hit.url,
        content: highlighter.highlightMarkdown(hit.content),
      });
    }

    return sorted;
  },
};

export default function DefaultSearchDialog(props: SharedProps) {
  const { search, setSearch, query } = useDocsSearch({ client: searchClient });

  return (
    <SearchDialog search={search} onSearchChange={setSearch} isLoading={query.isLoading} {...props}>
      <SearchDialogOverlay />
      <SearchDialogContent>
        <SearchDialogHeader>
          <SearchDialogIcon />
          <SearchDialogInput />
          <SearchDialogClose />
        </SearchDialogHeader>
        <SearchDialogList items={query.data !== 'empty' ? query.data : null} />
      </SearchDialogContent>
    </SearchDialog>
  );
}
