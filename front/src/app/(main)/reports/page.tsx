import { Suspense } from 'react';
import Link from 'next/link';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import { serverGet } from '@/utils/serverApi';
import { ReportSummary } from '@/types/models';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { formatSignedCurrency, formatPercent, formatDate } from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

async function ReportsList() {
  let data: ReportSummary[];
  try {
    data = await serverGet<ReportSummary[]>('/api/reports');
  } catch (e) {
    return <ErrorAlert message={e instanceof Error ? e.message : '取得に失敗しました'} />;
  }
  if (data.length === 0) {
    return (
      <EmptyState message="まだレポートがありません。毎週日曜18:00に自動生成されます。" />
    );
  }
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      {data.map((r) => (
        <Card key={r.week_start}>
          <CardContent>
            <Typography variant="subtitle1" fontWeight={600}>
              {formatDate(r.week_start)} 〜 {formatDate(r.week_end)}
            </Typography>
            <Box sx={{ display: 'flex', gap: 4, mt: 1, flexWrap: 'wrap' }}>
              <Typography variant="body2">取引数: {r.trade_count}</Typography>
              <Typography variant="body2">勝率: {formatPercent(r.win_rate)}</Typography>
              <Typography variant="body2" sx={{ color: pnlColor(r.total_pnl) }}>
                損益: {formatSignedCurrency(r.total_pnl)}
              </Typography>
            </Box>
          </CardContent>
          <CardActions>
            <Link href={`/reports/${r.week_start}`} style={{ textDecoration: 'none' }}>
              <Button size="small">詳細を見る</Button>
            </Link>
          </CardActions>
        </Card>
      ))}
    </Box>
  );
}

// 週次レポート一覧（RSC）。
export default function ReportsPage() {
  return (
    <Box>
      <PageTitle title="週次レポート" />
      <Suspense fallback={<LoadingSkeleton rows={3} />}>
        <ReportsList />
      </Suspense>
    </Box>
  );
}
