import { exportSearchIndexes } from '@/lib/export-search-indexes';

export const revalidate = false;

export function GET() {
  return Response.json(exportSearchIndexes());
}
