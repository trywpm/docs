import fs from 'node:fs';
import path from 'node:path';
import { readFile } from 'node:fs/promises';

import { algoliasearch } from 'algoliasearch';

import { algoliaIndexName, algoliaAppId } from '@/lib/shared';

if (!process.env.ALGOLIA_ADMIN_KEY) {
  console.error('ALGOLIA_ADMIN_KEY environment variable is not set.');
  process.exit(1);
}

const index = path.join(process.cwd(), 'out/docs/static.json');
if (!fs.existsSync(index)) {
  console.error(`Index file not found, run "pnpm run build" first`);
  process.exit(1);
}

const data = JSON.parse(await readFile(index, 'utf-8'));
if (!data || !Array.isArray(data)) {
  console.error(`Invalid index data in file: ${index}`);
  process.exit(1);
}

if (data.length === 0) {
  console.warn(`No records found in index file: ${index}`);
  process.exit(0);
}

const client = algoliasearch(algoliaAppId, process.env.ALGOLIA_ADMIN_KEY);

const lps = await client.replaceAllObjects({
  objects: data,
  indexName: algoliaIndexName,
});
console.warn(lps);
