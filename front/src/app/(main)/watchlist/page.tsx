'use client';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Alert from '@mui/material/Alert';
import { useFetch } from '@/hooks/useFetch';
import { WatchlistItem } from '@/types/models';
import {
  PageTitle,
  EmptyState,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { formatCurrency, formatSignedPercent } from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

const modeLabel: Record<string, string> = {
  virtual: 'バーチャル',
  real: '実運用',
  both: '両方',
};

export default function WatchlistPage() {
  const { data, loading, error } = useFetch<WatchlistItem[]>('/api/watchlist');

  return (
    <Box>
      <PageTitle title="ウォッチリスト" />
      <Alert severity="info" sx={{ mb: 3 }}>
        銘柄の追加・変更は管理者が行います。
      </Alert>

      {loading ? (
        <LoadingSkeleton rows={3} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : data && data.length > 0 ? (
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          {data.map((w) => (
            <Card key={w.id} sx={{ minWidth: 240, flex: '1 1 240px' }}>
              <CardContent>
                <Box
                  sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}
                >
                  <Box>
                    <Typography fontWeight={600}>{w.name ?? w.ticker}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {w.ticker}
                    </Typography>
                  </Box>
                  <Chip label={modeLabel[w.mode] ?? w.mode} size="small" variant="outlined" />
                </Box>
                <Divider sx={{ my: 1.5 }} />
                <Typography variant="h5">{formatCurrency(w.close)}</Typography>
                <Typography variant="body1" sx={{ color: pnlColor(w.change_rate) }}>
                  {formatSignedPercent(w.change_rate)}
                </Typography>
              </CardContent>
            </Card>
          ))}
        </Box>
      ) : (
        <EmptyState message="ウォッチリストに銘柄がありません" />
      )}
    </Box>
  );
}
