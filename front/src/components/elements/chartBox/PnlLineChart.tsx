'use client';

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { formatCurrency } from '@/utils/format';

export type PnlPoint = { date: string; cumulative_pnl: number };

// 累計損益の折れ線グラフ（X=日付・Y=累計損益）。
export const PnlLineChart = ({ data }: { data: PnlPoint[] }) => {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <LineChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#eee" />
        <XAxis dataKey="date" fontSize={12} />
        <YAxis
          fontSize={12}
          tickFormatter={(v: number) => formatCurrency(v)}
          width={80}
        />
        <Tooltip
          formatter={(v: number) => [formatCurrency(v), '累計損益']}
          labelFormatter={(l: string) => `日付: ${l}`}
        />
        <Line
          type="monotone"
          dataKey="cumulative_pnl"
          stroke="#1976d2"
          strokeWidth={2}
          dot={{ r: 3 }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
};
