'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
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
import AddIcon from '@mui/icons-material/Add';
import { apiClient } from '@/utils/apiClient';
import { WatchlistItem } from '@/types/models';
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
import { formatCurrency, formatSignedPercent } from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const addSchema = z.object({
  code: z.string().regex(/^[0-9]{4}$/, '銘柄コードは4桁の数字で入力してください'),
  mode: z.enum(['virtual', 'real', 'both']),
});
type AddValues = z.infer<typeof addSchema>;

const modeLabel: Record<string, string> = { virtual: 'バーチャル', real: '実運用', both: '両方' };

export default function AdminWatchlistPage() {
  const { data, loading, error, reload } = useFetch<WatchlistItem[]>('/api/watchlist');
  const { showSuccess, showError, ToastView } = useToast();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<WatchlistItem | null>(null);
  const [deleting, setDeleting] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<AddValues>({ resolver: zodResolver(addSchema), defaultValues: { mode: 'both' } });

  const onAdd = async (v: AddValues) => {
    try {
      await apiClient.post('/api/admin/watchlist', { code: v.code, mode: v.mode });
      showSuccess('銘柄を追加しました');
      setDialogOpen(false);
      reset({ code: '', mode: 'both' });
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '追加に失敗しました');
    }
  };

  const onDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await apiClient.delete(`/api/admin/watchlist/${deleteTarget.id}`);
      showSuccess('銘柄を削除しました');
      setDeleteTarget(null);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '削除に失敗しました');
    } finally {
      setDeleting(false);
    }
  };

  const count = data?.length ?? 0;

  return (
    <Box>
      <PageTitle
        title="ウォッチリスト管理"
        action={
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            disabled={count >= 3}
            onClick={() => setDialogOpen(true)}
          >
            銘柄を追加
          </Button>
        }
      />

      {count >= 3 && (
        <Box sx={{ mb: 2, color: 'text.secondary' }}>
          登録上限（3銘柄）に達しています。追加するには既存銘柄を削除してください。
        </Box>
      )}

      {loading ? (
        <LoadingSkeleton rows={3} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : data && data.length > 0 ? (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>銘柄</TableCell>
                <TableCell>モード</TableCell>
                <TableCell align="right">現在値</TableCell>
                <TableCell align="right">前日比</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data.map((w) => (
                <TableRow key={w.id}>
                  <TableCell>
                    {w.name ?? w.ticker}
                    <Box component="span" sx={{ color: 'text.secondary', ml: 1 }}>
                      {w.ticker}
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip label={modeLabel[w.mode] ?? w.mode} size="small" variant="outlined" />
                  </TableCell>
                  <TableCell align="right">{formatCurrency(w.close)}</TableCell>
                  <TableCell align="right" sx={{ color: pnlColor(w.change_rate) }}>
                    {formatSignedPercent(w.change_rate)}
                  </TableCell>
                  <TableCell align="right">
                    <Button size="small" color="error" onClick={() => setDeleteTarget(w)}>
                      削除
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <EmptyState message="ウォッチリストに銘柄がありません" />
      )}

      {/* 追加ダイアログ */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>銘柄を追加</DialogTitle>
        <form onSubmit={handleSubmit(onAdd)}>
          <DialogContent>
            <TextField
              label="銘柄コード（4桁）"
              fullWidth
              margin="normal"
              placeholder="7203"
              error={!!errors.code}
              helperText={errors.code?.message}
              {...register('code')}
            />
            <TextField
              select
              label="モード"
              fullWidth
              margin="normal"
              defaultValue="both"
              error={!!errors.mode}
              helperText={errors.mode?.message}
              {...register('mode')}
            >
              <MenuItem value="virtual">バーチャル</MenuItem>
              <MenuItem value="real">実運用</MenuItem>
              <MenuItem value="both">両方</MenuItem>
            </TextField>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDialogOpen(false)}>キャンセル</Button>
            <LoadingButton label="追加" type="submit" loading={isSubmitting} />
          </DialogActions>
        </form>
      </Dialog>

      <ConfirmDialog
        open={!!deleteTarget}
        title="銘柄の削除"
        message={`${deleteTarget?.name ?? deleteTarget?.ticker} を削除しますか？`}
        confirmLabel="削除"
        confirmColor="error"
        loading={deleting}
        onConfirm={onDelete}
        onClose={() => setDeleteTarget(null)}
      />

      {ToastView}
    </Box>
  );
}
