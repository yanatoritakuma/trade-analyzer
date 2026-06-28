'use client';

import { useState } from 'react';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import { apiClient } from '@/utils/apiClient';
import { WatchlistCandidate } from '@/types/models';
import { useFetch } from '@/hooks/useFetch';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { LoadingButton } from '@/components/elements/buttonBox/LoadingButton';
import { useToast } from '@/components/common/useToast';
import { formatDateTime } from '@/utils/format';

const statusChip: Record<string, { label: string; color: 'success' | 'error' | 'warning' }> = {
  pending: { label: '承認待ち', color: 'warning' },
  approved: { label: '承認済', color: 'success' },
  rejected: { label: '却下', color: 'error' },
};

export default function WatchlistCandidatesPage() {
  const { data, loading, error, reload } = useFetch<WatchlistCandidate[]>(
    '/api/admin/watchlist-candidates'
  );
  const { showSuccess, showError, ToastView } = useToast();

  const [approveTarget, setApproveTarget] = useState<WatchlistCandidate | null>(null);
  const [mode, setMode] = useState('both');
  const [busy, setBusy] = useState(false);
  const [rejectingId, setRejectingId] = useState<number | null>(null);

  const pending = data?.filter((c) => c.status === 'pending') ?? [];
  const history = data?.filter((c) => c.status !== 'pending') ?? [];

  const onApprove = async () => {
    if (!approveTarget) return;
    setBusy(true);
    try {
      await apiClient.patch(`/api/admin/watchlist-candidates/${approveTarget.id}/approve`, {
        mode,
      });
      showSuccess('候補を承認しました');
      setApproveTarget(null);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '承認に失敗しました');
    } finally {
      setBusy(false);
    }
  };

  const onReject = async (c: WatchlistCandidate) => {
    setRejectingId(c.id);
    try {
      await apiClient.patch(`/api/admin/watchlist-candidates/${c.id}/reject`);
      showSuccess('候補を却下しました');
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '却下に失敗しました');
    } finally {
      setRejectingId(null);
    }
  };

  return (
    <Box>
      <PageTitle title="ウォッチリスト候補承認" />

      {loading ? (
        <LoadingSkeleton rows={4} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : (
        <>
          <Typography variant="h6" gutterBottom>
            承認待ち
          </Typography>
          {pending.length > 0 ? (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mb: 4 }}>
              {pending.map((c) => (
                <Card key={c.id}>
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                      <Typography fontWeight={600}>
                        {c.name ?? c.ticker}（{c.ticker}）
                      </Typography>
                      {c.confidence != null && (
                        <Chip
                          label={`信頼度 ${Math.round(c.confidence * 100)}%`}
                          size="small"
                          variant="outlined"
                        />
                      )}
                      {c.replace_ticker && (
                        <Chip label={`置換: ${c.replace_ticker}`} size="small" color="warning" />
                      )}
                    </Box>
                    <Typography variant="body2" color="text.secondary">
                      {c.reason ?? '理由なし'}
                    </Typography>
                  </CardContent>
                  <CardActions>
                    <Button
                      color="success"
                      onClick={() => {
                        setMode('both');
                        setApproveTarget(c);
                      }}
                    >
                      承認
                    </Button>
                    <LoadingButton
                      label="却下"
                      variant="text"
                      color="error"
                      loading={rejectingId === c.id}
                      onClick={() => onReject(c)}
                    />
                  </CardActions>
                </Card>
              ))}
            </Box>
          ) : (
            <EmptyState message="承認待ちの候補はありません" />
          )}

          <Typography variant="h6" gutterBottom>
            承認・却下履歴
          </Typography>
          {history.length > 0 ? (
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>銘柄</TableCell>
                    <TableCell>状態</TableCell>
                    <TableCell>判断日時</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {history.map((c) => {
                    const chip = statusChip[c.status];
                    return (
                      <TableRow key={c.id}>
                        <TableCell>
                          {c.name ?? c.ticker}（{c.ticker}）
                        </TableCell>
                        <TableCell>
                          <Chip label={chip.label} size="small" color={chip.color} />
                        </TableCell>
                        <TableCell>{formatDateTime(c.decided_at)}</TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <EmptyState message="履歴はありません" />
          )}
        </>
      )}

      {/* 承認ダイアログ */}
      <Dialog open={!!approveTarget} onClose={() => setApproveTarget(null)} maxWidth="xs" fullWidth>
        <DialogTitle>候補の承認</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            {approveTarget?.name ?? approveTarget?.ticker} をウォッチリストに追加します。
            {approveTarget?.replace_ticker
              ? `（${approveTarget.replace_ticker} を置き換えます）`
              : ''}
          </DialogContentText>
          <TextField
            select
            label="モード"
            fullWidth
            value={mode}
            onChange={(e) => setMode(e.target.value)}
          >
            <MenuItem value="virtual">バーチャル</MenuItem>
            <MenuItem value="real">実運用</MenuItem>
            <MenuItem value="both">両方</MenuItem>
          </TextField>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setApproveTarget(null)}>キャンセル</Button>
          <LoadingButton label="承認する" color="success" loading={busy} onClick={onApprove} />
        </DialogActions>
      </Dialog>

      {ToastView}
    </Box>
  );
}
