'use client';

import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import { ReactNode } from 'react';

type Props = {
  label: ReactNode;
  loading?: boolean;
  onClick?: () => void;
  type?: 'button' | 'submit';
  variant?: 'contained' | 'outlined' | 'text';
  color?: 'primary' | 'success' | 'error' | 'inherit';
  disabled?: boolean;
  fullWidth?: boolean;
  size?: 'small' | 'medium' | 'large';
};

// ローディング状態を管理する汎用ボタン（二重送信防止）。
export const LoadingButton = ({
  label,
  loading = false,
  onClick,
  type = 'button',
  variant = 'contained',
  color = 'primary',
  disabled = false,
  fullWidth = false,
  size = 'medium',
}: Props) => {
  return (
    <Button
      type={type}
      variant={variant}
      color={color}
      onClick={onClick}
      disabled={disabled || loading}
      fullWidth={fullWidth}
      size={size}
    >
      {loading ? <CircularProgress size={20} color="inherit" /> : label}
    </Button>
  );
};
