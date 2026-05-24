import { algoliasearch } from 'algoliasearch';

import { algoliaIndexName, algoliaAppId } from '@/lib/shared';
import { exportSearchIndexes } from '@/lib/export-search-indexes';

if (!process.env.ALGOLIA_ADMIN_KEY) {
  console.error('ALGOLIA_ADMIN_KEY environment variable is not set.');
  process.exit(1);
}

const client = algoliasearch(algoliaAppId, process.env.ALGOLIA_ADMIN_KEY);

await client.replaceAllObjects({
  indexName: algoliaIndexName,
  // @ts-expect-error - BaseIndex satisfies the type expected by Algolia.
  objects: exportSearchIndexes(),
});
