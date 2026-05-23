'use client';

import { FrameworkProvider } from 'fumadocs-core/framework';
import { ViewOptionsPopover } from 'fumadocs-ui/layouts/docs/page';
import { useParams, usePathname as useNextPathname, useRouter } from 'next/navigation';

import { docsPathPrefix } from '@/lib/shared';

function useBasePathPathname() {
  const pathname = useNextPathname();
  return pathname === '/' ? docsPathPrefix : `${docsPathPrefix}${pathname}`;
}

export function ViewOptions(props: { markdownUrl: string; githubUrl: string }) {
  return (
    <FrameworkProvider
      useRouter={useRouter}
      useParams={useParams}
      usePathname={useBasePathPathname}
    >
      <ViewOptionsPopover {...props} />
    </FrameworkProvider>
  );
}
