'use client';

import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import { LoadingButton } from '../buttonBox/LoadingButton';

type Props = {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  confirmColor?: 'primary' | 'error';
  loading?: boolean;
  onConfirm: () => void;
  onClose: () => void;
};

// 確認ダイアログ（削除・停止などの確認用）。
export const ConfirmDialog = ({
  open,
  title,
  message,
  confirmLabel = 'OK',
  confirmColor = 'primary',
  loading = false,
  onConfirm,
  onClose,
}: Props) => {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <DialogContentText>{message}</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          キャンセル
        </Button>
        <LoadingButton
          label={confirmLabel}
          color={confirmColor}
          loading={loading}
          onClick={onConfirm}
        />
      </DialogActions>
    </Dialog>
  );
};
