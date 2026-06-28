// 損益値に応じた色（緑=プラス / 赤=マイナス / グレー=未確定）。
// Server / Client 双方から呼べるよう、'use client' を持たない純粋関数モジュールに置く。
export const pnlColor = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return '#9e9e9e';
  if (value > 0) return '#2e7d32';
  if (value < 0) return '#c62828';
  return '#616161';
};
