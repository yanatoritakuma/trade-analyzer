'use client';

import { useState, useCallback } from 'react';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';

type Severity = 'success' | 'error' | 'info';

// 成功/失敗トーストを簡単に出すフック。
export const useToast = () => {
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState('');
  const [severity, setSeverity] = useState<Severity>('success');

  const show = useCallback((msg: string, sev: Severity = 'success') => {
    setMessage(msg);
    setSeverity(sev);
    setOpen(true);
  }, []);

  const showSuccess = useCallback((msg: string) => show(msg, 'success'), [show]);
  const showError = useCallback((msg: string) => show(msg, 'error'), [show]);

  const ToastView = (
    <Snackbar
      open={open}
      autoHideDuration={4000}
      onClose={() => setOpen(false)}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
    >
      <Alert
        severity={severity}
        variant="filled"
        onClose={() => setOpen(false)}
        sx={{ width: '100%' }}
      >
        {message}
      </Alert>
    </Snackbar>
  );

  return { showSuccess, showError, ToastView };
};
