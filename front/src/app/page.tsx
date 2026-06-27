'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/utils/apiClient';

type HealthResponse = {
  status: string;
  db: string;
};

export default function Home() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiClient
      .get<HealthResponse>('/health')
      .then(setHealth)
      .catch((e: Error) => setError(e.message));
  }, []);

  return (
    <main style={{ padding: 32, fontFamily: 'sans-serif' }}>
      <h1>{process.env.NEXT_PUBLIC_APP_NAME ?? 'AI Trading System'}</h1>
      <p>開発環境セットアップ確認用ページ</p>

      <section style={{ marginTop: 24 }}>
        <h2>バックエンド接続状態</h2>
        {error && <p style={{ color: '#c62828' }}>接続エラー: {error}</p>}
        {!error && !health && <p>確認中...</p>}
        {health && (
          <p style={{ color: '#2e7d32' }}>
            status: {health.status} / db: {health.db}
          </p>
        )}
      </section>
    </main>
  );
}
