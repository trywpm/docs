import { join } from 'node:path';
import { existsSync } from 'node:fs';
import { cp, mkdir, readdir, rename, rm } from 'node:fs/promises';

const OUT_DIR = 'out';
const NESTED_DIR = join(OUT_DIR, 'docs');
const NOT_FOUND_FILE = '404.html';

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
    if (entry === 'docs') {
      continue;
    }
    await rename(join(OUT_DIR, entry), join(NESTED_DIR, entry));
  }

  // Cloudflare Workers Assets reads `404.html` at the root of the assets
  // directory for `not_found_handling: "404-page"`. Mirror it alongside the
  // nested copy so the not-found path resolves even though every real page
  // lives under /docs.
  const nested404 = join(NESTED_DIR, NOT_FOUND_FILE);
  if (existsSync(nested404)) {
    await cp(nested404, join(OUT_DIR, NOT_FOUND_FILE));
  }
}

await main();
