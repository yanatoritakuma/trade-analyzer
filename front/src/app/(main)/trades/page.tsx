'use client';

import { useState, useEffect, useCallback, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableFooter from '@mui/material/TableFooter';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import { apiClient } from '@/utils/apiClient';
import { Trade, TradeListResponse, TradeMode, WatchlistItem } from '@/types/models';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import {
  formatCurrency,
  formatSignedCurrency,
  formatDate,
  formatPercent,
  formatNumber,
} from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const TradeDetailModal = ({
  trade,
  onClose,
}: {
  trade: Trade | null;
  onClose: () => void;
}) => {
  if (!trade) return null;
  return (
    <Dialog open={!!trade} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        {trade.name ?? trade.ticker}（{trade.ticker}）
      </DialogTitle>
      <DialogContent dividers>
        <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
          <Chip
            label={trade.action}
            color={trade.action === 'BUY' ? 'success' : 'error'}
            size="small"
          />
          <Chip label={`数量 ${formatNumber(trade.quantity)}`} size="small" variant="outlined" />
          <Chip label={`単価 ${formatCurrency(trade.price)}`} size="small" variant="outlined" />
        </Box>

        <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', mb: 2 }}>
          <Box>
            <Typography variant="caption" color="text.secondary">
              目標株価
            </Typography>
            <Typography>{formatCurrency(trade.target_price)}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              損切りライン
            </Typography>
            <Typography>{formatCurrency(trade.stop_loss)}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              信頼度
            </Typography>
            <Typography>
              {trade.confidence != null ? `${Math.round(trade.confidence * 100)}%` : '—'}
            </Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              損益
            </Typography>
            <Typography sx={{ color: pnlColor(trade.result_pnl) }}>
              {trade.result_pnl != null ? formatSignedCurrency(trade.result_pnl) : '未確定'}
            </Typography>
          </Box>
        </Box>

        {trade.reason && (
          <>
            <Typography variant="subtitle2" fontWeight={600}>
              総合コメント
            </Typography>
            <Typography variant="body2" sx={{ mb: 2 }}>
              {trade.reason}
            </Typography>
          </>
        )}

        {trade.buy_reasons && trade.buy_reasons.length > 0 && (
          <>
            <Divider sx={{ my: 1 }} />
            <Typography variant="subtitle2" fontWeight={600} color="success.main">
              買う理由
            </Typography>
            <ul style={{ marginTop: 4 }}>
              {trade.buy_reasons.map((r, i) => (
                <li key={i}>
                  <Typography variant="body2">{r}</Typography>
                </li>
              ))}
            </ul>
          </>
        )}

        {trade.no_buy_reasons && trade.no_buy_reasons.length > 0 && (
          <>
            <Typography variant="subtitle2" fontWeight={600} color="error.main">
              買わない理由
            </Typography>
            <ul style={{ marginTop: 4 }}>
              {trade.no_buy_reasons.map((r, i) => (
                <li key={i}>
                  <Typography variant="body2">{r}</Typography>
                </li>
              ))}
            </ul>
          </>
        )}

        {trade.entry_condition && (
          <>
            <Typography variant="subtitle2" fontWeight={600}>
              エントリー条件
            </Typography>
            <Typography variant="body2">{trade.entry_condition}</Typography>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};

function TradesContent() {
  const params = useSearchParams();
  const [mode, setMode] = useState<TradeMode>('virtual');
  const [ticker, setTicker] = useState(params.get('ticker') ?? '');
  const [action, setAction] = useState('');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [data, setData] = useState<TradeListResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<Trade | null>(null);
  const [tickers, setTickers] = useState<string[]>([]);

  // 銘柄フィルタの選択肢はウォッチリスト由来（dev_spec_05）
  useEffect(() => {
    apiClient
      .get<WatchlistItem[]>('/api/watchlist')
      .then((ws) => setTickers(ws.map((w) => w.ticker)))
      .catch(() => setTickers([]));
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const q: Record<string, string> = { mode };
      if (ticker) q.ticker = ticker;
      if (action) q.action = action;
      if (from) q.from = from;
      if (to) q.to = to;
      const res = await apiClient.get<TradeListResponse>('/api/trades', q);
      setData(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : '取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, [mode, ticker, action, from, to]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <Box>
      <PageTitle title="トレード履歴" />

      <Tabs value={mode} onChange={(_, v) => setMode(v)} sx={{ mb: 2 }}>
        <Tab label="バーチャル" value="virtual" />
        <Tab label="実運用" value="real" />
      </Tabs>

      <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mb: 2 }}>
        <TextField
          select
          label="銘柄"
          size="small"
          value={ticker}
          onChange={(e) => setTicker(e.target.value)}
          sx={{ minWidth: 160 }}
        >
          <MenuItem value="">すべて</MenuItem>
          {Array.from(new Set([...tickers, ...(ticker ? [ticker] : [])])).map((t) => (
            <MenuItem key={t} value={t}>
              {t}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          select
          label="売買"
          size="small"
          value={action}
          onChange={(e) => setAction(e.target.value)}
          sx={{ minWidth: 120 }}
        >
          <MenuItem value="">すべて</MenuItem>
          <MenuItem value="BUY">BUY</MenuItem>
          <MenuItem value="SELL">SELL</MenuItem>
        </TextField>
        <TextField
          label="開始日"
          type="date"
          size="small"
          value={from}
          onChange={(e) => setFrom(e.target.value)}
          InputLabelProps={{ shrink: true }}
        />
        <TextField
          label="終了日"
          type="date"
          size="small"
          value={to}
          onChange={(e) => setTo(e.target.value)}
          InputLabelProps={{ shrink: true }}
        />
      </Box>

      {loading ? (
        <LoadingSkeleton rows={5} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : data && data.items.length > 0 ? (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>日付</TableCell>
                <TableCell>銘柄</TableCell>
                <TableCell>売買</TableCell>
                <TableCell align="right">単価</TableCell>
                <TableCell align="right">数量</TableCell>
                <TableCell align="right">損益</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data.items.map((t) => (
                <TableRow
                  key={t.id}
                  hover
                  sx={{ cursor: 'pointer' }}
                  onClick={() => setSelected(t)}
                >
                  <TableCell>{formatDate(t.created_at)}</TableCell>
                  <TableCell>{t.name ?? t.ticker}</TableCell>
                  <TableCell>
                    <Chip
                      label={t.action}
                      size="small"
                      color={t.action === 'BUY' ? 'success' : 'error'}
                    />
                  </TableCell>
                  <TableCell align="right">{formatCurrency(t.price)}</TableCell>
                  <TableCell align="right">{formatNumber(t.quantity)}</TableCell>
                  <TableCell align="right" sx={{ color: pnlColor(t.result_pnl) }}>
                    {t.result_pnl != null ? formatSignedCurrency(t.result_pnl) : '未確定'}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell colSpan={2}>
                  <Typography variant="body2" fontWeight={600}>
                    件数: {data.summary.count}
                  </Typography>
                </TableCell>
                <TableCell colSpan={2}>
                  <Typography variant="body2" fontWeight={600}>
                    勝率: {formatPercent(data.summary.win_rate)}
                  </Typography>
                </TableCell>
                <TableCell colSpan={2} align="right">
                  <Typography
                    variant="body2"
                    fontWeight={600}
                    sx={{ color: pnlColor(data.summary.total_pnl) }}
                  >
                    累計損益: {formatSignedCurrency(data.summary.total_pnl)}
                  </Typography>
                </TableCell>
              </TableRow>
            </TableFooter>
          </Table>
        </TableContainer>
      ) : (
        <EmptyState message="トレード履歴がありません" />
      )}

      <TradeDetailModal trade={selected} onClose={() => setSelected(null)} />
    </Box>
  );
}

export default function TradesPage() {
  return (
    <Suspense fallback={null}>
      <TradesContent />
    </Suspense>
  );
}
