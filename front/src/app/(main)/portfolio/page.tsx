'use client';

import { useMemo, useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import { apiClient } from '@/utils/apiClient';
import {
  PortfolioSummary,
  Position,
  TradeListResponse,
  TradeMode,
  ModeSummary,
} from '@/types/models';
import { PnlLineChart, PnlPoint } from '@/components/elements/chartBox/PnlLineChart';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import {
  formatCurrency,
  formatSignedCurrency,
  formatPercent,
  formatSignedPercent,
  formatDate,
  formatNumber,
} from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const periodMonths: Record<string, number> = { '1M': 1, '3M': 3, '6M': 6 };

export default function PortfolioPage() {
  const router = useRouter();
  const [mode, setMode] = useState<TradeMode>('virtual');
  const [period, setPeriod] = useState('3M');

  const [summary, setSummary] = useState<PortfolioSummary | null>(null);
  const [trades, setTrades] = useState<TradeListResponse | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // 損益推移グラフ用の決済データ（trades）と、保有ポジション（mode別・現在値/含み益付き）を取得
      const [s, t, p] = await Promise.all([
        apiClient.get<PortfolioSummary>('/api/portfolio/summary'),
        apiClient.get<TradeListResponse>('/api/trades', { mode }),
        apiClient.get<Position[]>('/api/positions', { mode }),
      ]);
      setSummary(s);
      setTrades(t);
      setPositions(p);
    } catch (e) {
      setError(e instanceof Error ? e.message : '取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, [mode]);

  useEffect(() => {
    void load();
  }, [load]);

  const modeSummary: ModeSummary | undefined =
    mode === 'virtual' ? summary?.virtual : summary?.real;

  // 損益推移（決済トレードの累積）
  const chartData: PnlPoint[] = useMemo(() => {
    if (!trades) return [];
    const cutoff = new Date();
    cutoff.setMonth(cutoff.getMonth() - (periodMonths[period] ?? 3));
    const closed = trades.items
      .filter((t) => t.closed_at && t.result_pnl != null)
      .filter((t) => new Date(t.closed_at as string) >= cutoff)
      .sort(
        (a, b) =>
          new Date(a.closed_at as string).getTime() -
          new Date(b.closed_at as string).getTime()
      );
    let cumulative = 0;
    return closed.map((t) => {
      cumulative += t.result_pnl as number;
      return { date: formatDate(t.closed_at), cumulative_pnl: cumulative };
    });
  }, [trades, period]);

  return (
    <Box>
      <PageTitle title="ポートフォリオ" />

      <Tabs value={mode} onChange={(_, v) => setMode(v)} sx={{ mb: 2 }}>
        <Tab label="バーチャル" value="virtual" />
        <Tab label="実運用" value="real" />
      </Tabs>

      {loading ? (
        <LoadingSkeleton rows={6} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : (
        <>
          {/* サマリーカード */}
          {modeSummary && (
            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mb: 3 }}>
              <Card sx={{ flex: 1, minWidth: 180 }}>
                <CardContent>
                  <Typography variant="caption" color="text.secondary">
                    累計損益
                  </Typography>
                  <Typography variant="h6" sx={{ color: pnlColor(modeSummary.total_pnl) }}>
                    {formatSignedCurrency(modeSummary.total_pnl)}
                  </Typography>
                </CardContent>
              </Card>
              <Card sx={{ flex: 1, minWidth: 180 }}>
                <CardContent>
                  <Typography variant="caption" color="text.secondary">
                    勝率
                  </Typography>
                  <Typography variant="h6">{formatPercent(modeSummary.win_rate)}</Typography>
                </CardContent>
              </Card>
              <Card sx={{ flex: 1, minWidth: 180 }}>
                <CardContent>
                  <Typography variant="caption" color="text.secondary">
                    トレード数
                  </Typography>
                  <Typography variant="h6">{modeSummary.trade_count}</Typography>
                </CardContent>
              </Card>
            </Box>
          )}

          {/* 損益推移グラフ */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Box
                sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}
              >
                <Typography variant="h6">損益推移</Typography>
                <ToggleButtonGroup
                  size="small"
                  exclusive
                  value={period}
                  onChange={(_, v) => v && setPeriod(v)}
                >
                  <ToggleButton value="1M">1ヶ月</ToggleButton>
                  <ToggleButton value="3M">3ヶ月</ToggleButton>
                  <ToggleButton value="6M">6ヶ月</ToggleButton>
                </ToggleButtonGroup>
              </Box>
              {chartData.length > 0 ? (
                <PnlLineChart data={chartData} />
              ) : (
                <EmptyState message="表示できる決済データがありません" />
              )}
            </CardContent>
          </Card>

          {/* 保有ポジション */}
          <Typography variant="h6" gutterBottom>
            保有ポジション
          </Typography>
          {positions.length > 0 ? (
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>銘柄</TableCell>
                    <TableCell align="right">数量</TableCell>
                    <TableCell align="right">取得単価</TableCell>
                    <TableCell align="right">現在値</TableCell>
                    <TableCell align="right">含み益</TableCell>
                    <TableCell align="right">損益率</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {positions.map((p) => (
                    <TableRow
                      key={p.ticker}
                      hover
                      sx={{ cursor: 'pointer' }}
                      onClick={() => router.push(`/trades?ticker=${p.ticker}`)}
                    >
                      <TableCell>{p.name ?? p.ticker}</TableCell>
                      <TableCell align="right">{formatNumber(p.quantity)}</TableCell>
                      <TableCell align="right">{formatCurrency(p.avg_price)}</TableCell>
                      <TableCell align="right">{formatCurrency(p.close)}</TableCell>
                      <TableCell align="right" sx={{ color: pnlColor(p.unrealized_pnl) }}>
                        {p.unrealized_pnl != null ? formatSignedCurrency(p.unrealized_pnl) : '—'}
                      </TableCell>
                      <TableCell align="right" sx={{ color: pnlColor(p.pnl_rate) }}>
                        {p.pnl_rate != null ? formatSignedPercent(p.pnl_rate) : '—'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <EmptyState message="現在保有中のポジションはありません" />
          )}
        </>
      )}
    </Box>
  );
}
