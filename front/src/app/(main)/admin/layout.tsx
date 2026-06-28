'use client';

import { ReactNode } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';

const ADMIN_TABS = [
  { label: 'ダッシュボード', path: '/admin' },
  { label: 'ユーザー', path: '/admin/users' },
  { label: '招待コード', path: '/admin/invitations' },
  { label: 'ウォッチリスト', path: '/admin/watchlist' },
  { label: '分析設定', path: '/admin/analysis-settings' },
  { label: '候補承認', path: '/admin/watchlist-candidates' },
];

// 現在パスに対応するタブ値を求める（サブページは親タブをアクティブに）。
const resolveTab = (pathname: string): string => {
  if (pathname.startsWith('/admin/analysis-settings')) return '/admin/analysis-settings';
  const match = ADMIN_TABS.find((t) => t.path === pathname);
  if (match) return match.path;
  return '/admin';
};

export default function AdminLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  return (
    <Box>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs
          value={resolveTab(pathname)}
          onChange={(_, v) => router.push(v)}
          variant="scrollable"
          scrollButtons="auto"
        >
          {ADMIN_TABS.map((t) => (
            <Tab key={t.path} label={t.label} value={t.path} />
          ))}
        </Tabs>
      </Box>
      {children}
    </Box>
  );
}
