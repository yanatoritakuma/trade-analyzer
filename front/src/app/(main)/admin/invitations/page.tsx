'use client';

import { useState } from 'react';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import AddIcon from '@mui/icons-material/Add';
import { apiClient } from '@/utils/apiClient';
import { Invitation } from '@/types/models';
import { useFetch } from '@/hooks/useFetch';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { ConfirmDialog } from '@/components/elements/modalBox/ConfirmDialog';
import { LoadingButton } from '@/components/elements/buttonBox/LoadingButton';
import { useToast } from '@/components/common/useToast';
import { formatDateTime, formatDate } from '@/utils/format';

const statusChip: Record<string, { label: string; color: 'success' | 'default' | 'warning' | 'error' }> = {
  valid: { label: '有効', color: 'success' },
  used: { label: '使用済', color: 'default' },
  expired: { label: '期限切', color: 'warning' },
  disabled: { label: '無効', color: 'error' },
};

export default function AdminInvitationsPage() {
  const { data, loading, error, reload } = useFetch<Invitation[]>('/api/admin/invitations');
  const { showSuccess, showError, ToastView } = useToast();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [expiresDays, setExpiresDays] = useState(7);
  const [creating, setCreating] = useState(false);
  const [disableTarget, setDisableTarget] = useState<Invitation | null>(null);
  const [disabling, setDisabling] = useState(false);

  const onCreate = async () => {
    setCreating(true);
    try {
      const res = await apiClient.post<{ code: string; expires_at: string }>(
        '/api/admin/invitations',
        { expires_days: expiresDays }
      );
      try {
        await navigator.clipboard.writeText(res.code);
        showSuccess(`招待コード ${res.code} を発行し、クリップボードにコピーしました`);
      } catch {
        showSuccess(`招待コード ${res.code} を発行しました`);
      }
      setDialogOpen(false);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '発行に失敗しました');
    } finally {
      setCreating(false);
    }
  };

  const onDisable = async () => {
    if (!disableTarget) return;
    setDisabling(true);
    try {
      await apiClient.delete(`/api/admin/invitations/${disableTarget.id}`);
      showSuccess('招待コードを無効化しました');
      setDisableTarget(null);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '無効化に失敗しました');
    } finally {
      setDisabling(false);
    }
  };

  const copyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      showSuccess('コピーしました');
    } catch {
      showError('コピーに失敗しました');
    }
  };

  return (
    <Box>
      <PageTitle
        title="招待コード管理"
        action={
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setDialogOpen(true)}>
            招待コードを発行
          </Button>
        }
      />

      {loading ? (
        <LoadingSkeleton rows={4} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : data && data.length > 0 ? (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>コード</TableCell>
                <TableCell>状態</TableCell>
                <TableCell>有効期限</TableCell>
                <TableCell>発行日</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data.map((inv) => {
                const chip = statusChip[inv.status] ?? statusChip.disabled;
                return (
                  <TableRow key={inv.id}>
                    <TableCell>
                      {inv.code}
                      <IconButton size="small" onClick={() => copyCode(inv.code)}>
                        <ContentCopyIcon fontSize="inherit" />
                      </IconButton>
                    </TableCell>
                    <TableCell>
                      <Chip label={chip.label} size="small" color={chip.color} />
                    </TableCell>
                    <TableCell>{formatDateTime(inv.expires_at)}</TableCell>
                    <TableCell>{formatDate(inv.created_at)}</TableCell>
                    <TableCell align="right">
                      {inv.status === 'valid' && (
                        <Button size="small" color="error" onClick={() => setDisableTarget(inv)}>
                          無効化
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <EmptyState message="招待コードがありません" />
      )}

      {/* 発行ダイアログ */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>招待コードを発行</DialogTitle>
        <DialogContent>
          <TextField
            select
            label="有効期限"
            fullWidth
            margin="normal"
            value={expiresDays}
            onChange={(e) => setExpiresDays(Number(e.target.value))}
          >
            {[3, 7, 14, 30].map((d) => (
              <MenuItem key={d} value={d}>
                {d}日
              </MenuItem>
            ))}
          </TextField>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>キャンセル</Button>
          <LoadingButton label="発行する" loading={creating} onClick={onCreate} />
        </DialogActions>
      </Dialog>

      <ConfirmDialog
        open={!!disableTarget}
        title="招待コードの無効化"
        message={`${disableTarget?.code} を無効化しますか？`}
        confirmLabel="無効化"
        confirmColor="error"
        loading={disabling}
        onConfirm={onDisable}
        onClose={() => setDisableTarget(null)}
      />

      {ToastView}
    </Box>
  );
}
