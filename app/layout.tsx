import type { Metadata } from 'next';

import { Banner } from 'fumadocs-ui/components/banner';

import './global.css';
import { appName, siteUrl } from '@/lib/shared';
import { Provider } from '@/components/provider';

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  icons: [
    {
      url: 'https://wpm.so/favicon.svg',
      type: 'image/svg+xml',
    },
  ],
  title: {
    default: `wpm - Modern package management for WordPress`,
    template: `%s - ${appName}`,
  },
  description: 'Discover, install, and manage WordPress packages like never before.',
  applicationName: appName,
  openGraph: {
    title: {
      default: `wpm - Modern package management for WordPress`,
      template: `%s - ${appName}`,
    },
    siteName: appName,
    type: 'website',
    locale: 'en_US',
  },
  twitter: {
    card: 'summary_large_image',
    title: {
      default: `wpm - Modern package management for WordPress`,
      template: `%s - ${appName}`,
    },
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html lang="en" data-scroll-behavior="smooth" suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <Banner id="wpm-status-banner" variant="normal">
          wpm is still under development. Expect breaking changes as we work toward a stable
          release.
        </Banner>
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
