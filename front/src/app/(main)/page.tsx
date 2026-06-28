import { Suspense } from 'react';
import Link from 'next/link';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import { serverGet } from '@/utils/serverApi';
import {
  PortfolioSummary,
  AnalysisSignal,
  WatchlistItem,
  ModeSummary,
  SignalAction,
} from '@/types/models';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import {
  formatSignedCurrency,
  formatPercent,
  formatSignedPercent,
  formatCurrency,
} from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const signalColor: Record<SignalAction, 'success' | 'error' | 'default'> = {
  BUY: 'success',
  SELL: 'error',
  HOLD: 'default',
};

const cardLinkStyle = { textDecoration: 'none' } as const;

const SummaryCard = ({ title, s }: { title: string; s: ModeSummary }) => (
  <Card sx={{ flex: 1, minWidth: 240 }}>
    <CardContent>
      <Typography variant="subtitle1" fontWeight={600} gutterBottom>
        {title}
      </Typography>
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
        <Box>
          <Typography variant="caption" color="text.secondary">
            累計損益
          </Typography>
          <Typography variant="h6" sx={{ color: pnlColor(s.total_pnl) }}>
            {formatSignedCurrency(s.total_pnl)}
          </Typography>
        </Box>
        <Box>
          <Typography variant="caption" color="text.secondary">
            今週損益
          </Typography>
          <Typography variant="h6" sx={{ color: pnlColor(s.weekly_pnl) }}>
            {formatSignedCurrency(s.weekly_pnl)}
          </Typography>
        </Box>
        <Box>
          <Typography variant="caption" color="text.secondary">
            勝率
          </Typography>
          <Typography variant="h6">{formatPercent(s.win_rate)}</Typography>
        </Box>
        <Box>
          <Typography variant="caption" color="text.secondary">
            トレード数
          </Typography>
          <Typography variant="h6">{s.trade_count}</Typography>
        </Box>
      </Box>
    </CardContent>
  </Card>
);

async function SummarySection() {
  let data: PortfolioSummary;
  try {
    data = await serverGet<PortfolioSummary>('/api/portfolio/summary');
  } catch (e) {
    return <ErrorAlert message={e instanceof Error ? e.message : '取得に失敗しました'} />;
  }
  return (
    <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mb: 4 }}>
      <SummaryCard title="バーチャル" s={data.virtual} />
      <SummaryCard title="実運用" s={data.real} />
    </Box>
  );
}

async function SignalsSection() {
  let data: AnalysisSignal[];
  try {
    data = await serverGet<AnalysisSignal[]>('/api/analysis/latest', { limit: '3' });
  } catch (e) {
    return <ErrorAlert message={e instanceof Error ? e.message : '取得に失敗しました'} />;
  }
  if (data.length === 0) return <EmptyState message="まだ分析データがありません" />;
  return (
    <Box sx={{ mb: 4 }}>
      {data.map((sig, i) => (
        <Link key={i} href="/trades" style={cardLinkStyle}>
          <Card sx={{ mb: 1, '&:hover': { boxShadow: 3 } }}>
            <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1.5 }}>
              <Chip label={sig.action} color={signalColor[sig.action]} size="small" />
              <Typography sx={{ flexGrow: 1 }} color="text.primary">
                {sig.name ?? sig.ticker}（{sig.ticker}）
              </Typography>
              <Typography variant="body2" color="text.secondary">
                信頼度 {sig.confidence != null ? `${Math.round(sig.confidence * 100)}%` : '—'}
              </Typography>
            </CardContent>
          </Card>
        </Link>
      ))}
    </Box>
  );
}

async function WatchlistSection() {
  let data: WatchlistItem[];
  try {
    data = await serverGet<WatchlistItem[]>('/api/watchlist');
  } catch (e) {
    return <ErrorAlert message={e instanceof Error ? e.message : '取得に失敗しました'} />;
  }
  if (data.length === 0) return <EmptyState message="ウォッチリストに銘柄がありません" />;
  return (
    <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
      {data.map((w) => (
        <Link
          key={w.id}
          href="/watchlist"
          style={{ ...cardLinkStyle, flex: '1 1 220px', minWidth: 220 }}
        >
          <Card sx={{ height: '100%', '&:hover': { boxShadow: 3 } }}>
            <CardContent>
              <Typography fontWeight={600} color="text.primary">
                {w.name ?? w.ticker}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {w.ticker}
              </Typography>
              <Divider sx={{ my: 1 }} />
              <Typography variant="h6" color="text.primary">
                {formatCurrency(w.close)}
              </Typography>
              <Typography variant="body2" sx={{ color: pnlColor(w.change_rate) }}>
                {formatSignedPercent(w.change_rate)}
              </Typography>
            </CardContent>
          </Card>
        </Link>
      ))}
    </Box>
  );
}

// ダッシュボード（RSC）。各セクションは Suspense で並列ストリーミングする。
export default function DashboardPage() {
  return (
    <Box>
      <PageTitle title="ダッシュボード" />

      <Typography variant="h6" gutterBottom>
        損益サマリー
      </Typography>
      <Suspense fallback={<LoadingSkeleton rows={2} />}>
        <SummarySection />
      </Suspense>

      <Typography variant="h6" gutterBottom>
        最新分析シグナル
      </Typography>
      <Suspense fallback={<LoadingSkeleton rows={3} />}>
        <SignalsSection />
      </Suspense>

      <Typography variant="h6" gutterBottom>
        ウォッチリスト
      </Typography>
      <Suspense fallback={<LoadingSkeleton rows={3} />}>
        <WatchlistSection />
      </Suspense>
    </Box>
  );
}
