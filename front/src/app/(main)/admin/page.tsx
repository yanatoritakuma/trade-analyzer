'use client';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { useFetch } from '@/hooks/useFetch';
import {
  AdminUser,
  Invitation,
  AnalysisSignal,
  PortfolioSummary,
} from '@/types/models';
import { PageTitle, LoadingSkeleton, ErrorAlert } from '@/components/common/StateView';
import {
  formatSignedCurrency,
  formatPercent,
  formatDateTime,
} from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const StatCard = ({ label, value }: { label: string; value: string }) => (
  <Card sx={{ flex: 1, minWidth: 180 }}>
    <CardContent>
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="h5">{value}</Typography>
    </CardContent>
  </Card>
);

export default function AdminDashboardPage() {
  const users = useFetch<AdminUser[]>('/api/admin/users');
  const invitations = useFetch<Invitation[]>('/api/admin/invitations');
  const signals = useFetch<AnalysisSignal[]>('/api/analysis/latest', { limit: '1' });
  const summary = useFetch<PortfolioSummary>('/api/portfolio/summary');

  const validInvites = invitations.data?.filter((i) => i.status === 'valid').length ?? 0;
  const lastAnalysis = signals.data?.[0]?.analyzed_at ?? null;

  const loading = users.loading || invitations.loading || signals.loading || summary.loading;
  const error = users.error || invitations.error || summary.error;

  return (
    <Box>
      <PageTitle title="管理ダッシュボード" />

      {loading ? (
        <LoadingSkeleton rows={4} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : (
        <>
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mb: 4 }}>
            <StatCard label="登録ユーザー数" value={String(users.data?.length ?? 0)} />
            <StatCard label="有効な招待コード" value={String(validInvites)} />
            <StatCard
              label="最終分析日時"
              value={lastAnalysis ? formatDateTime(lastAnalysis) : '—'}
            />
          </Box>

          <Typography variant="h6" gutterBottom>
            共通ポートフォリオ成績
          </Typography>
          {summary.data && (
            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mb: 4 }}>
              {(['virtual', 'real'] as const).map((m) => {
                const s = summary.data![m];
                return (
                  <Card key={m} sx={{ flex: 1, minWidth: 240 }}>
                    <CardContent>
                      <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                        {m === 'virtual' ? 'バーチャル' : '実運用'}
                      </Typography>
                      <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
                        <Box>
                          <Typography variant="caption" color="text.secondary">
                            勝率
                          </Typography>
                          <Typography>{formatPercent(s.win_rate)}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">
                            今週損益
                          </Typography>
                          <Typography sx={{ color: pnlColor(s.weekly_pnl) }}>
                            {formatSignedCurrency(s.weekly_pnl)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">
                            累計損益
                          </Typography>
                          <Typography sx={{ color: pnlColor(s.total_pnl) }}>
                            {formatSignedCurrency(s.total_pnl)}
                          </Typography>
                        </Box>
                      </Box>
                    </CardContent>
                  </Card>
                );
              })}
            </Box>
          )}

          <Typography variant="h6" gutterBottom>
            ユーザー一覧
          </Typography>
          <Paper>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>名前</TableCell>
                  <TableCell>ロール</TableCell>
                  <TableCell>状態</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {users.data?.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell>{u.name}</TableCell>
                    <TableCell>{u.role === 'admin' ? '管理者' : '一般'}</TableCell>
                    <TableCell>
                      <Chip
                        label={u.is_active ? '有効' : '停止中'}
                        size="small"
                        color={u.is_active ? 'success' : 'default'}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Paper>
        </>
      )}
    </Box>
  );
}
