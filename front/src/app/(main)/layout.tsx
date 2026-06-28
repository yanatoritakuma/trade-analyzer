import { ReactNode } from 'react';
import { AppShell } from '@/components/layout/AppShell';

// 認証必須ページ共通レイアウト（Navbar + Sidebar）。
export default function MainLayout({ children }: { children: ReactNode }) {
  return <AppShell>{children}</AppShell>;
}
