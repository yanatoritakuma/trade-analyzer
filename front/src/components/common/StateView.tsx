'use client';

import { ReactNode } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import Skeleton from '@mui/material/Skeleton';

// ページ見出し。
export const PageTitle = ({
  title,
  action,
}: {
  title: string;
  action?: ReactNode;
}) => (
  <Box
    sx={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      mb: 3,
    }}
  >
    <Typography variant="h5" component="h1" fontWeight={600}>
      {title}
    </Typography>
    {action}
  </Box>
);

// 空状態の表示。
export const EmptyState = ({ message }: { message: string }) => (
  <Box sx={{ textAlign: 'center', py: 6, color: 'text.secondary' }}>
    <Typography>{message}</Typography>
  </Box>
);

// エラー表示。
export const ErrorAlert = ({ message }: { message: string }) => (
  <Alert severity="error" sx={{ my: 2 }}>
    {message}
  </Alert>
);

// 読み込み中のスケルトン（行数指定）。
export const LoadingSkeleton = ({ rows = 4 }: { rows?: number }) => (
  <Box>
    {Array.from({ length: rows }).map((_, i) => (
      <Skeleton key={i} height={48} sx={{ my: 1 }} />
    ))}
  </Box>
);
