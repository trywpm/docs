import { join } from 'node:path';
import { existsSync } from 'node:fs';
import { cp, mkdir, readdir, rename, rm, writeFile } from 'node:fs/promises';

const OUT_DIR = 'out';
const NESTED_DIR = join(OUT_DIR, 'docs');
const NOT_FOUND_FILE = '404.html';
const NOT_FOUND_URL = 'https://wpm.so/404';
const ASSETS_ROOT_FILES = new Set(['_headers', '_redirects']);

async function fetchUpstream404(): Promise<string> {
  const res = await fetch(NOT_FOUND_URL);
  const body = await res.text();
  if (!body) {
    throw new Error(`[nest-export] ${NOT_FOUND_URL} returned an empty body.`);
  }
  return body;
}

async function main(): Promise<void> {
  if (!existsSync(OUT_DIR)) {
    console.error(`[nest-export] ${OUT_DIR}/ does not exist — run \`next build\` first.`);
    process.exit(1);
  }

  if (existsSync(NESTED_DIR)) {
    await rm(NESTED_DIR, { recursive: true, force: true });
  }
  await mkdir(NESTED_DIR, { recursive: true });

  const entries: string[] = await readdir(OUT_DIR);
  for (const entry of entries) {
    if (entry === 'docs' || ASSETS_ROOT_FILES.has(entry)) {
      continue;
    }
    await rename(join(OUT_DIR, entry), join(NESTED_DIR, entry));
  }

  // Replace the Next-generated 404 with the main site's 404 so docs stays
  // visually consistent with wpm.so.
  const nested404 = join(NESTED_DIR, NOT_FOUND_FILE);
  await writeFile(nested404, await fetchUpstream404(), 'utf8');

  // Cloudflare Workers Assets reads `404.html` at the root of the assets
  // directory for `not_found_handling: "404-page"`. Mirror it alongside the
  // nested copy so the not-found path resolves even though every real page
  // lives under /docs.
  await cp(nested404, join(OUT_DIR, NOT_FOUND_FILE));
}

await main();
