import { format } from 'date-fns';

// 金額を「¥1,234」形式にフォーマットする（小数は四捨五入）。
export const formatCurrency = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return '—';
  const sign = value < 0 ? '-' : '';
  return `${sign}¥${Math.abs(Math.round(value)).toLocaleString('ja-JP')}`;
};

// 損益を符号付きで表示する（+¥1,234 / -¥1,234）。
export const formatSignedCurrency = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return '—';
  const sign = value > 0 ? '+' : value < 0 ? '-' : '';
  return `${sign}¥${Math.abs(Math.round(value)).toLocaleString('ja-JP')}`;
};

// パーセントを「12.3%」形式にフォーマットする。
export const formatPercent = (value: number | null | undefined, digits = 1): string => {
  if (value === null || value === undefined) return '—';
  return `${value.toFixed(digits)}%`;
};

// 符号付きパーセント（+1.8% / -1.5%）。
export const formatSignedPercent = (value: number | null | undefined, digits = 2): string => {
  if (value === null || value === undefined) return '—';
  const sign = value > 0 ? '+' : '';
  return `${sign}${value.toFixed(digits)}%`;
};

// ISO日時を「2026/06/27」形式に整形する。
export const formatDate = (value: string | null | undefined): string => {
  if (!value) return '—';
  try {
    return format(new Date(value), 'yyyy/MM/dd');
  } catch {
    return '—';
  }
};

// ISO日時を「2026/06/27 15:30」形式に整形する。
export const formatDateTime = (value: string | null | undefined): string => {
  if (!value) return '—';
  try {
    return format(new Date(value), 'yyyy/MM/dd HH:mm');
  } catch {
    return '—';
  }
};

// 数値（株数など）を桁区切りで表示する。
export const formatNumber = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return '—';
  return value.toLocaleString('ja-JP');
};
