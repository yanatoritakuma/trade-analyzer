'use client';

import { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Box from '@mui/material/Box';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import IconButton from '@mui/material/IconButton';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import { apiClient } from '@/utils/apiClient';
import { useAuth } from '@/context/AuthContext';
import { Position } from '@/types/models';
import { PageTitle } from '@/components/common/StateView';
import { LoadingButton } from '@/components/elements/buttonBox/LoadingButton';
import { ConfirmDialog } from '@/components/elements/modalBox/ConfirmDialog';
import { useToast } from '@/components/common/useToast';
import { formatCurrency, formatNumber } from '@/utils/format';

// ---- プロフィール ----
const profileSchema = z.object({
  name: z.string().min(1, 'お名前を入力してください').max(50, '50文字以内で入力してください'),
});
type ProfileValues = z.infer<typeof profileSchema>;

// ---- パスワード ----
const passwordSchema = z
  .object({
    currentPassword: z.string().min(1, '現在のパスワードを入力してください'),
    newPassword: z
      .string()
      .min(8, '8文字以上で入力してください')
      .regex(/^(?=.*[a-zA-Z])(?=.*[0-9])/, '英字と数字を含めてください'),
    confirmPassword: z.string().min(1, '確認用パスワードを入力してください'),
  })
  .refine((d) => d.newPassword === d.confirmPassword, {
    message: 'パスワードが一致しません',
    path: ['confirmPassword'],
  });
type PasswordValues = z.infer<typeof passwordSchema>;

// ---- 保有株 ----
const positionSchema = z.object({
  code: z.string().regex(/^[0-9]{4}$/, '銘柄コードは4桁の数字で入力してください'),
  avgPrice: z.coerce.number().min(1, '1以上の数値を入力してください'),
  quantity: z.coerce.number().int().min(1, '1以上の整数を入力してください'),
});
type PositionValues = z.infer<typeof positionSchema>;

export default function SettingsPage() {
  const { user, reload } = useAuth();
  const { showSuccess, showError, ToastView } = useToast();

  const profileForm = useForm<ProfileValues>({ resolver: zodResolver(profileSchema) });
  const passwordForm = useForm<PasswordValues>({ resolver: zodResolver(passwordSchema) });

  useEffect(() => {
    if (user) profileForm.reset({ name: user.name });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  const onProfile = async (v: ProfileValues) => {
    try {
      await apiClient.patch('/api/admin/me', { name: v.name });
      await reload();
      showSuccess('プロフィールを更新しました');
    } catch (e) {
      showError(e instanceof Error ? e.message : '更新に失敗しました');
    }
  };

  const onPassword = async (v: PasswordValues) => {
    try {
      await apiClient.put('/api/admin/me/password', {
        current_password: v.currentPassword,
        new_password: v.newPassword,
      });
      passwordForm.reset({ currentPassword: '', newPassword: '', confirmPassword: '' });
      showSuccess('パスワードを変更しました');
    } catch (e) {
      showError(e instanceof Error ? e.message : '変更に失敗しました');
    }
  };

  // ---- 保有株 ----
  const [positions, setPositions] = useState<Position[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Position | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Position | null>(null);
  const [deleting, setDeleting] = useState(false);

  const positionForm = useForm<PositionValues>({ resolver: zodResolver(positionSchema) });

  const loadPositions = useCallback(async () => {
    try {
      const p = await apiClient.get<Position[]>('/api/positions');
      setPositions(p);
    } catch {
      /* noop */
    }
  }, []);

  useEffect(() => {
    void loadPositions();
  }, [loadPositions]);

  const openCreate = () => {
    setEditing(null);
    positionForm.reset({ code: '', avgPrice: 0, quantity: 0 });
    setDialogOpen(true);
  };

  const openEdit = (p: Position) => {
    setEditing(p);
    positionForm.reset({
      code: p.ticker.replace('.T', ''),
      avgPrice: p.avg_price,
      quantity: p.quantity,
    });
    setDialogOpen(true);
  };

  const onSavePosition = async (v: PositionValues) => {
    try {
      const body = { code: v.code, avg_price: v.avgPrice, quantity: v.quantity };
      if (editing) {
        await apiClient.put(`/api/admin/positions/${editing.id}`, body);
        showSuccess('保有株を更新しました');
      } else {
        await apiClient.post('/api/admin/positions', body);
        showSuccess('保有株を登録しました');
      }
      setDialogOpen(false);
      await loadPositions();
    } catch (e) {
      showError(e instanceof Error ? e.message : '保存に失敗しました');
    }
  };

  const onDeletePosition = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await apiClient.delete(`/api/admin/positions/${deleteTarget.id}`);
      showSuccess('保有株を削除しました');
      setDeleteTarget(null);
      await loadPositions();
    } catch (e) {
      showError(e instanceof Error ? e.message : '削除に失敗しました');
    } finally {
      setDeleting(false);
    }
  };

  return (
    <Box>
      <PageTitle title="設定" />

      {/* プロフィール */}
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <Typography fontWeight={600}>プロフィール</Typography>
        </AccordionSummary>
        <AccordionDetails>
          <form onSubmit={profileForm.handleSubmit(onProfile)}>
            <TextField
              label="メールアドレス"
              value={user?.email ?? ''}
              fullWidth
              margin="normal"
              disabled
            />
            <TextField
              label="お名前"
              fullWidth
              margin="normal"
              error={!!profileForm.formState.errors.name}
              helperText={profileForm.formState.errors.name?.message}
              {...profileForm.register('name')}
            />
            <Box mt={2}>
              <LoadingButton
                label="更新する"
                type="submit"
                loading={profileForm.formState.isSubmitting}
              />
            </Box>
          </form>
        </AccordionDetails>
      </Accordion>

      {/* パスワード変更 */}
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <Typography fontWeight={600}>パスワード変更</Typography>
        </AccordionSummary>
        <AccordionDetails>
          <form onSubmit={passwordForm.handleSubmit(onPassword)}>
            <TextField
              label="現在のパスワード"
              type="password"
              fullWidth
              margin="normal"
              error={!!passwordForm.formState.errors.currentPassword}
              helperText={passwordForm.formState.errors.currentPassword?.message}
              {...passwordForm.register('currentPassword')}
            />
            <TextField
              label="新しいパスワード"
              type="password"
              fullWidth
              margin="normal"
              error={!!passwordForm.formState.errors.newPassword}
              helperText={passwordForm.formState.errors.newPassword?.message}
              {...passwordForm.register('newPassword')}
            />
            <TextField
              label="新しいパスワード（確認）"
              type="password"
              fullWidth
              margin="normal"
              error={!!passwordForm.formState.errors.confirmPassword}
              helperText={passwordForm.formState.errors.confirmPassword?.message}
              {...passwordForm.register('confirmPassword')}
            />
            <Box mt={2}>
              <LoadingButton
                label="変更する"
                type="submit"
                loading={passwordForm.formState.isSubmitting}
              />
            </Box>
          </form>
        </AccordionDetails>
      </Accordion>

      {/* 保有株 */}
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <Typography fontWeight={600}>実運用 保有株</Typography>
        </AccordionSummary>
        <AccordionDetails>
          <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1 }}>
            <Button startIcon={<AddIcon />} variant="outlined" onClick={openCreate}>
              保有株を追加
            </Button>
          </Box>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>銘柄</TableCell>
                <TableCell align="right">取得単価</TableCell>
                <TableCell align="right">数量</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {positions.map((p) => (
                <TableRow key={p.id}>
                  <TableCell>{p.name ?? p.ticker}</TableCell>
                  <TableCell align="right">{formatCurrency(p.avg_price)}</TableCell>
                  <TableCell align="right">{formatNumber(p.quantity)}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => openEdit(p)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => setDeleteTarget(p)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
              {positions.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} align="center" sx={{ color: 'text.secondary' }}>
                    保有株が登録されていません
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </AccordionDetails>
      </Accordion>

      {/* 保有株ダイアログ */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>{editing ? '保有株を編集' : '保有株を追加'}</DialogTitle>
        <form onSubmit={positionForm.handleSubmit(onSavePosition)}>
          <DialogContent>
            <TextField
              label="銘柄コード（4桁）"
              fullWidth
              margin="normal"
              placeholder="7203"
              error={!!positionForm.formState.errors.code}
              helperText={positionForm.formState.errors.code?.message}
              {...positionForm.register('code')}
            />
            <TextField
              label="取得単価"
              type="number"
              fullWidth
              margin="normal"
              error={!!positionForm.formState.errors.avgPrice}
              helperText={positionForm.formState.errors.avgPrice?.message}
              {...positionForm.register('avgPrice')}
            />
            <TextField
              label="数量"
              type="number"
              fullWidth
              margin="normal"
              error={!!positionForm.formState.errors.quantity}
              helperText={positionForm.formState.errors.quantity?.message}
              {...positionForm.register('quantity')}
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDialogOpen(false)}>キャンセル</Button>
            <LoadingButton
              label="保存"
              type="submit"
              loading={positionForm.formState.isSubmitting}
            />
          </DialogActions>
        </form>
      </Dialog>

      <ConfirmDialog
        open={!!deleteTarget}
        title="保有株の削除"
        message={`${deleteTarget?.name ?? deleteTarget?.ticker} を削除しますか？`}
        confirmLabel="削除"
        confirmColor="error"
        loading={deleting}
        onConfirm={onDeletePosition}
        onClose={() => setDeleteTarget(null)}
      />

      {ToastView}
    </Box>
  );
}
