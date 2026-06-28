'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import { apiClient } from '@/utils/apiClient';
import { AnalysisSetting, AnalysisStyle, Theme } from '@/types/models';
import {
  PageTitle,
  ErrorAlert,
  LoadingSkeleton,
} from '@/components/common/StateView';
import { LoadingButton } from '@/components/elements/buttonBox/LoadingButton';
import { useToast } from '@/components/common/useToast';

type Period = 'short_term' | 'mid_term';
type Direction = 'trend' | 'contrarian' | 'both';

const toStyle = (p: Period, d: Direction): AnalysisStyle =>
  `${p}_${d}` as AnalysisStyle;

const fromStyle = (style: AnalysisStyle | undefined): { period: Period; direction: Direction } => {
  // style例: short_term_trend / mid_term_both
  if (!style) return { period: 'short_term', direction: 'trend' };
  if (style.startsWith('short_term')) {
    return { period: 'short_term', direction: style.replace('short_term_', '') as Direction };
  }
  return { period: 'mid_term', direction: style.replace('mid_term_', '') as Direction };
};

export default function AnalysisSettingsPage() {
  const router = useRouter();
  const { showSuccess, showError, ToastView } = useToast();

  const [themes, setThemes] = useState<Theme[]>([]);
  const [selectedThemes, setSelectedThemes] = useState<number[]>([]);
  const [period, setPeriod] = useState<Period>('short_term');
  const [direction, setDirection] = useState<Direction>('trend');
  const [minMarketCap, setMinMarketCap] = useState('');
  const [minVolume, setMinVolume] = useState('');
  const [maxPer, setMaxPer] = useState('');
  const [freePrompt, setFreePrompt] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [themeList, setting] = await Promise.all([
        apiClient.get<Theme[]>('/api/admin/analysis-themes'),
        apiClient.get<AnalysisSetting>('/api/admin/analysis-settings'),
      ]);
      setThemes(themeList);
      setSelectedThemes(setting.theme_ids ?? []);
      const { period: p, direction: d } = fromStyle(setting.style);
      setPeriod(p);
      setDirection(d);
      if (setting.screening) {
        setMinMarketCap(setting.screening.min_market_cap ? String(setting.screening.min_market_cap) : '');
        setMinVolume(setting.screening.min_volume ? String(setting.screening.min_volume) : '');
        setMaxPer(setting.screening.max_per ? String(setting.screening.max_per) : '');
      }
      setFreePrompt(setting.free_prompt ?? '');
    } catch (e) {
      setError(e instanceof Error ? e.message : '取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const toggleTheme = (id: number) => {
    setSelectedThemes((prev) =>
      prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id]
    );
  };

  const onSave = async () => {
    if (selectedThemes.length === 0) {
      showError('テーマを1つ以上選択してください');
      return;
    }
    if (freePrompt.length > 1000) {
      showError('自由プロンプトは1000文字以内で入力してください');
      return;
    }
    setSaving(true);
    try {
      await apiClient.put('/api/admin/analysis-settings', {
        theme_ids: selectedThemes,
        style: toStyle(period, direction),
        screening: {
          min_market_cap: Number(minMarketCap) || 0,
          min_volume: Number(minVolume) || 0,
          max_per: Number(maxPer) || 0,
        },
        free_prompt: freePrompt,
      });
      showSuccess('分析設定を保存しました。次回15:30から反映されます。');
    } catch (e) {
      showError(e instanceof Error ? e.message : '保存に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Box>
      <PageTitle
        title="分析設定"
        action={
          <Button variant="outlined" onClick={() => router.push('/admin/analysis-settings/themes')}>
            テーマ管理
          </Button>
        }
      />

      {loading ? (
        <LoadingSkeleton rows={5} />
      ) : error ? (
        <ErrorAlert message={error} />
      ) : (
        <>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                分析テーマ（1つ以上選択）
              </Typography>
              {themes.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                  テーマが登録されていません。「テーマ管理」から追加してください。
                </Typography>
              ) : (
                <FormGroup>
                  {themes.map((t) => (
                    <FormControlLabel
                      key={t.id}
                      control={
                        <Checkbox
                          checked={selectedThemes.includes(t.id)}
                          onChange={() => toggleTheme(t.id)}
                        />
                      }
                      label={t.name}
                    />
                  ))}
                </FormGroup>
              )}
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                分析スタイル
              </Typography>
              <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                <TextField
                  select
                  label="期間"
                  value={period}
                  onChange={(e) => setPeriod(e.target.value as Period)}
                  sx={{ minWidth: 160 }}
                >
                  <MenuItem value="short_term">短期</MenuItem>
                  <MenuItem value="mid_term">中期</MenuItem>
                </TextField>
                <TextField
                  select
                  label="方向"
                  value={direction}
                  onChange={(e) => setDirection(e.target.value as Direction)}
                  sx={{ minWidth: 160 }}
                >
                  <MenuItem value="trend">順張り</MenuItem>
                  <MenuItem value="contrarian">逆張り</MenuItem>
                  <MenuItem value="both">両方</MenuItem>
                </TextField>
              </Box>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                スクリーニング条件
              </Typography>
              <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                <TextField
                  label="時価総額（円以上）"
                  type="number"
                  value={minMarketCap}
                  onChange={(e) => setMinMarketCap(e.target.value)}
                />
                <TextField
                  label="出来高（株以上）"
                  type="number"
                  value={minVolume}
                  onChange={(e) => setMinVolume(e.target.value)}
                />
                <TextField
                  label="PER（以下）"
                  type="number"
                  value={maxPer}
                  onChange={(e) => setMaxPer(e.target.value)}
                />
              </Box>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                自由プロンプト（最大1000文字）
              </Typography>
              <TextField
                multiline
                minRows={3}
                fullWidth
                value={freePrompt}
                onChange={(e) => setFreePrompt(e.target.value)}
                helperText={`${freePrompt.length} / 1000`}
                error={freePrompt.length > 1000}
              />
              <Divider sx={{ my: 2 }} />
              <Typography variant="caption" color="text.secondary">
                プレビュー: テーマ {selectedThemes.length} 件 / スタイル {toStyle(period, direction)}
              </Typography>
            </CardContent>
          </Card>

          <LoadingButton label="保存する" loading={saving} onClick={onSave} />
        </>
      )}

      {ToastView}
    </Box>
  );
}
