import jsonSchema from 'fumadocs-mdx/plugins/json-schema';
import lastModified from 'fumadocs-mdx/plugins/last-modified';
import { defineConfig, defineDocs } from 'fumadocs-mdx/config';
import { metaSchema, pageSchema } from 'fumadocs-core/source/schema';
import { remarkSteps } from 'fumadocs-core/mdx-plugins';

export const docs = defineDocs({
  dir: 'content',
  docs: {
    schema: pageSchema,
    postprocess: {
      includeProcessedMarkdown: true,
      extractLinkReferences: true,
      valueToExport: ['elementIds'],
    },
  },
  meta: {
    schema: metaSchema,
  },
});

export default defineConfig({
  plugins: [
    jsonSchema({
      insert: true,
    }),
    lastModified(),
  ],
  mdxOptions: {
    remarkPlugins: [remarkSteps],
  },
});
