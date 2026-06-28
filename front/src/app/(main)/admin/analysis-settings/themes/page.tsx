'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
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
import IconButton from '@mui/material/IconButton';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { apiClient } from '@/utils/apiClient';
import { Theme } from '@/types/models';
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

const themeSchema = z.object({
  name: z.string().min(1, 'テーマ名を入力してください').max(100, '100文字以内で入力してください'),
  description: z.string().max(255, '255文字以内で入力してください').optional(),
  isActive: z.boolean(),
});
type ThemeValues = z.infer<typeof themeSchema>;

export default function ThemesPage() {
  const router = useRouter();
  const { data, loading, error, reload } = useFetch<Theme[]>('/api/admin/analysis-themes');
  const { showSuccess, showError, ToastView } = useToast();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Theme | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Theme | null>(null);
  const [deleting, setDeleting] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<ThemeValues>({
    resolver: zodResolver(themeSchema),
    defaultValues: { name: '', description: '', isActive: true },
  });

  const openCreate = () => {
    setEditing(null);
    reset({ name: '', description: '', isActive: true });
    setDialogOpen(true);
  };

  const openEdit = (t: Theme) => {
    setEditing(t);
    reset({ name: t.name, description: t.description ?? '', isActive: t.is_active });
    setDialogOpen(true);
  };

  const onSave = async (v: ThemeValues) => {
    try {
      const body = { name: v.name, description: v.description ?? '', is_active: v.isActive };
      if (editing) {
        await apiClient.put(`/api/admin/analysis-themes/${editing.id}`, body);
        showSuccess('テーマを更新しました');
      } else {
        await apiClient.post('/api/admin/analysis-themes', body);
        showSuccess('テーマを追加しました');
      }
      setDialogOpen(false);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '保存に失敗しました');
    }
  };

  const onDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await apiClient.delete(`/api/admin/analysis-themes/${deleteTarget.id}`);
      showSuccess('テーマを削除しました');
      setDeleteTarget(null);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '削除に失敗しました');
    } finally {
      setDeleting(false);
    }
  };

  // 並び替え（上下移動 → PATCH /sort）
  const move = async (index: number, dir: -1 | 1) => {
    if (!data) return;
    const target = index + dir;
    if (target < 0 || target >= data.length) return;
    const reordered = [...data];
    [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
    const payload = reordered.map((t, i) => ({ id: t.id, sort_order: i + 1 }));
    try {
      await apiClient.patch('/api/admin/analysis-themes/sort', payload);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '並び替えに失敗しました');
    }
  };

  return (
    <Box>
      <Button
        startIcon={<ArrowBackIcon />}
        onClick={() => router.push('/admin/analysis-settings')}
        sx={{ mb: 1 }}
      >
        分析設定に戻る
      </Button>
      <PageTitle
        title="テーマ管理"
        action={
          <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>
            テーマを追加
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
                <TableCell>並び</TableCell>
                <TableCell>テーマ名</TableCell>
                <TableCell>説明</TableCell>
                <TableCell>状態</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data.map((t, i) => (
                <TableRow key={t.id}>
                  <TableCell>
                    <IconButton size="small" disabled={i === 0} onClick={() => move(i, -1)}>
                      <ArrowUpwardIcon fontSize="inherit" />
                    </IconButton>
                    <IconButton
                      size="small"
                      disabled={i === data.length - 1}
                      onClick={() => move(i, 1)}
                    >
                      <ArrowDownwardIcon fontSize="inherit" />
                    </IconButton>
                  </TableCell>
                  <TableCell>{t.name}</TableCell>
                  <TableCell>{t.description ?? '—'}</TableCell>
                  <TableCell>
                    <Chip
                      label={t.is_active ? '有効' : '無効'}
                      size="small"
                      color={t.is_active ? 'success' : 'default'}
                    />
                  </TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => openEdit(t)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => setDeleteTarget(t)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <EmptyState message="テーマが登録されていません" />
      )}

      {/* 追加/編集ダイアログ */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>{editing ? 'テーマを編集' : 'テーマを追加'}</DialogTitle>
        <form onSubmit={handleSubmit(onSave)}>
          <DialogContent>
            <TextField
              label="テーマ名"
              fullWidth
              margin="normal"
              error={!!errors.name}
              helperText={errors.name?.message}
              {...register('name')}
            />
            <TextField
              label="説明"
              fullWidth
              margin="normal"
              multiline
              minRows={2}
              error={!!errors.description}
              helperText={errors.description?.message}
              {...register('description')}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={watch('isActive')}
                  onChange={(e) => setValue('isActive', e.target.checked)}
                />
              }
              label="有効"
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDialogOpen(false)}>キャンセル</Button>
            <LoadingButton label="保存" type="submit" loading={isSubmitting} />
          </DialogActions>
        </form>
      </Dialog>

      <ConfirmDialog
        open={!!deleteTarget}
        title="テーマの削除"
        message={`${deleteTarget?.name} を削除しますか？`}
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
