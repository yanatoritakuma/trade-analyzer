'use client';

import { createTheme } from '@mui/material/styles';

// アプリ共通のMUIテーマ。損益の緑/赤は palette.success / palette.error を使用する。
export const theme = createTheme({
  palette: {
    primary: { main: '#1976d2' },
    success: { main: '#2e7d32' },
    error: { main: '#c62828' },
    background: { default: '#f5f6f8' },
  },
  typography: {
    fontFamily: [
      'Hiragino Sans',
      'Hiragino Kaku Gothic ProN',
      'Meiryo',
      'sans-serif',
    ].join(','),
  },
  shape: { borderRadius: 8 },
});
