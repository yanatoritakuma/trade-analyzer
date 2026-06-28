import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Chip from '@mui/material/Chip';
import { Trade } from '@/types/models';
import {
  formatCurrency,
  formatSignedCurrency,
  formatDate,
  formatNumber,
} from '@/utils/format';
import { pnlColor } from '@/utils/pnl';

// 当週トレード一覧テーブル（読み取り専用・Server Componentから使用可）。
export const TradeMiniTable = ({ trades }: { trades: Trade[] }) => (
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
        {trades.map((t) => (
          <TableRow key={t.id}>
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
    </Table>
  </TableContainer>
);
