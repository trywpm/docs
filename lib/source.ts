import { icons } from 'lucide-react';
import { createElement } from 'react';
import { docs } from 'collections/server';
import { loader } from 'fumadocs-core/source';

import { docsContentRoute, docsImageRoute, docsRoute } from './shared';

function isLucideIcon(name: string): name is keyof typeof icons {
  return name in icons;
}

export const source = loader({
  source: docs.toFumadocsSource(),
  plugins: [],
  baseUrl: docsRoute,
  pageTree: {
    transformers: [
      {
        root(node) {
          const overview = node.children.find(
            (child) => child.type === 'page' && (child.url === '/' || child.url === ''),
          );
          const gettingStartedIdx = node.children.findIndex(
            (child) => child.type === 'folder' && child.name === 'Getting started',
          );
          if (!overview || gettingStartedIdx === -1) {
            return node;
          }

          const gettingStarted = node.children[gettingStartedIdx];
          if (gettingStarted.type !== 'folder') {
            return node;
          }

          const remaining = node.children.filter((child) => child !== overview);
          const idx = remaining.indexOf(gettingStarted);
          remaining[idx] = {
            ...gettingStarted,
            children: [overview, ...gettingStarted.children],
          };

          return { ...node, children: remaining };
        },
      },
    ],
  },
  icon(icon) {
    if (!icon || !isLucideIcon(icon)) {
      return null;
    }

    return createElement(icons[icon]);
  },
});

export function getPageImage(page: (typeof source)['$inferPage']) {
  const segments = [...page.slugs, 'image.png'];

  return {
    segments,
    url: `${docsImageRoute}/${segments.join('/')}`,
  };
}

export function getPageMarkdownUrl(page: (typeof source)['$inferPage']) {
  const segments = [...page.slugs, 'content.md'];

  return {
    segments,
    url: `${docsContentRoute}/${segments.join('/')}`,
  };
}

export async function getLLMText(page: (typeof source)['$inferPage']) {
  const processed = await page.data.getText('processed');

  return `# ${page.data.title} (${page.url})

${processed}`;
}
