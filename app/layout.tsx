import { Banner } from 'fumadocs-ui/components/banner';

import { Provider } from '@/components/provider';
import './global.css';

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <Banner id="wpm-status-banner" variant="normal">
          wpm is still under development. Expect breaking changes as we work
          toward a stable release.
        </Banner>
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
