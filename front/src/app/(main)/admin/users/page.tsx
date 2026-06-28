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
import { apiClient } from '@/utils/apiClient';
import { useAuth } from '@/context/AuthContext';
import { AdminUser } from '@/types/models';
import { useFetch } from '@/hooks/useFetch';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { ConfirmDialog } from '@/components/elements/modalBox/ConfirmDialog';
import { useToast } from '@/components/common/useToast';
import { formatDate } from '@/utils/format';

type Action = { type: 'toggle' | 'delete'; user: AdminUser } | null;

export default function AdminUsersPage() {
  const { user: me } = useAuth();
  const { data, loading, error, reload } = useFetch<AdminUser[]>('/api/admin/users');
  const { showSuccess, showError, ToastView } = useToast();
  const [action, setAction] = useState<Action>(null);
  const [busy, setBusy] = useState(false);

  const onConfirm = async () => {
    if (!action) return;
    setBusy(true);
    try {
      if (action.type === 'toggle') {
        await apiClient.patch(`/api/admin/users/${action.user.id}`, {
          is_active: !action.user.is_active,
        });
        showSuccess(action.user.is_active ? 'ユーザーを停止しました' : 'ユーザーを復活しました');
      } else {
        await apiClient.delete(`/api/admin/users/${action.user.id}`);
        showSuccess('ユーザーを削除しました');
      }
      setAction(null);
      reload();
    } catch (e) {
      showError(e instanceof Error ? e.message : '操作に失敗しました');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Box>
      <PageTitle title="ユーザー管理" />

      {loading ? (
        <LoadingSkeleton rows={4} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : data && data.length > 0 ? (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>名前</TableCell>
                <TableCell>メール</TableCell>
                <TableCell>ロール</TableCell>
                <TableCell>状態</TableCell>
                <TableCell>登録日</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data.map((u) => {
                const isSelfOrAdmin = u.role === 'admin' || u.id === me?.id;
                return (
                  <TableRow key={u.id}>
                    <TableCell>{u.name}</TableCell>
                    <TableCell>{u.email}</TableCell>
                    <TableCell>{u.role === 'admin' ? '管理者' : '一般'}</TableCell>
                    <TableCell>
                      <Chip
                        label={u.is_active ? '有効' : '停止中'}
                        size="small"
                        color={u.is_active ? 'success' : 'default'}
                      />
                    </TableCell>
                    <TableCell>{formatDate(u.created_at)}</TableCell>
                    <TableCell align="right">
                      {!isSelfOrAdmin && (
                        <>
                          <Button
                            size="small"
                            onClick={() => setAction({ type: 'toggle', user: u })}
                          >
                            {u.is_active ? '停止' : '復活'}
                          </Button>
                          <Button
                            size="small"
                            color="error"
                            onClick={() => setAction({ type: 'delete', user: u })}
                          >
                            削除
                          </Button>
                        </>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <EmptyState message="ユーザーがいません" />
      )}

      <ConfirmDialog
        open={!!action}
        title={action?.type === 'delete' ? 'ユーザーの削除' : 'ユーザーの状態変更'}
        message={
          action?.type === 'delete'
            ? `${action?.user.name} を削除しますか？この操作は取り消せません。`
            : `${action?.user.name} を${action?.user.is_active ? '停止' : '復活'}しますか？`
        }
        confirmLabel={action?.type === 'delete' ? '削除' : 'OK'}
        confirmColor={action?.type === 'delete' ? 'error' : 'primary'}
        loading={busy}
        onConfirm={onConfirm}
        onClose={() => setAction(null)}
      />

      {ToastView}
    </Box>
  );
}
