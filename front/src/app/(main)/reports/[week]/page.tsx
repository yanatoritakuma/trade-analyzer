import { Suspense } from 'react';
import Link from 'next/link';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { serverGet } from '@/utils/serverApi';
import { ReportDetail } from '@/types/models';
import {
  PageTitle,
  EmptyState,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { TradeMiniTable } from '@/components/reports/TradeMiniTable';
import {
  formatSignedCurrency,
  formatPercent,
  formatDate,
  formatCurrency,
} from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const MetricCard = ({ label, value, color }: { label: string; value: string; color?: string }) => (
  <Card sx={{ flex: 1, minWidth: 140 }}>
    <CardContent>
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="h6" sx={{ color }}>
        {value}
      </Typography>
    </CardContent>
  </Card>
);

const Section = ({ icon, title, body }: { icon: string; title: string; body: string }) => (
  <Card sx={{ mb: 2 }}>
    <CardContent>
      <Typography variant="subtitle1" fontWeight={600} gutterBottom>
        {icon} {title}
      </Typography>
      <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
        {body || '—'}
      </Typography>
    </CardContent>
  </Card>
);

async function ReportDetailContent({ week }: { week: string }) {
  let data: ReportDetail;
  try {
    data = await serverGet<ReportDetail>(`/api/reports/${week}`);
  } catch {
    return <EmptyState message="レポートが見つかりません" />;
  }
  return (
    <>
      <PageTitle
        title={`週次レポート（${formatDate(data.week_start)} 〜 ${formatDate(data.week_end)}）`}
      />

      <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mb: 3 }}>
        <MetricCard label="勝率" value={formatPercent(data.win_rate)} />
        <MetricCard
          label="損益"
          value={formatSignedCurrency(data.total_pnl)}
          color={pnlColor(data.total_pnl)}
        />
        <MetricCard label="取引数" value={String(data.trade_count)} />
        <MetricCard
          label="最大ドローダウン"
          value={data.max_drawdown != null ? formatCurrency(-data.max_drawdown) : '—'}
          color={data.max_drawdown ? pnlColor(-data.max_drawdown) : undefined}
        />
      </Box>

      <Typography variant="h6" gutterBottom>
        AI学習メモ
      </Typography>
      <Section icon="✅" title="有効だった点" body={data.lessons} />
      <Section icon="📝" title="サマリー" body={data.summary} />
      <Section icon="📌" title="来週の戦略" body={data.strategy} />

      <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
        当週のトレード
      </Typography>
      {data.trades.length > 0 ? (
        <TradeMiniTable trades={data.trades} />
      ) : (
        <EmptyState message="当週のトレードはありません" />
      )}
    </>
  );
}

// 週次レポート詳細（RSC）。
export default async function ReportDetailPage({
  params,
}: {
  params: Promise<{ week: string }>;
}) {
  const { week } = await params;
  return (
    <Box>
      <Link href="/reports" style={{ textDecoration: 'none' }}>
        <Button startIcon={<ArrowBackIcon />} sx={{ mb: 1 }}>
          一覧に戻る
        </Button>
      </Link>
      <Suspense fallback={<LoadingSkeleton rows={5} />}>
        <ReportDetailContent week={week} />
      </Suspense>
    </Box>
  );
}
