'use client';

import { ReactNode } from 'react';
import { AppRouterCacheProvider } from '@mui/material-nextjs/v15-appRouter';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { theme } from '@/theme';
import { AuthProvider } from '@/context/AuthContext';

// アプリ全体のクライアントプロバイダ（emotion SSRキャッシュ + MUIテーマ + 認証コンテキスト）。
// AppRouterCacheProvider により Server Component でのMUI描画もSSRスタイルが適用される。
export const Providers = ({ children }: { children: ReactNode }) => {
  return (
    <AppRouterCacheProvider options={{ key: 'mui' }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <AuthProvider>{children}</AuthProvider>
      </ThemeProvider>
    </AppRouterCacheProvider>
  );
};
