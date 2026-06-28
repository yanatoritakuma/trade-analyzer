import { cookies } from 'next/headers';

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080';

// serverGet は Server Component から、リクエストの access_token Cookie を転送して
// バックエンドGETを呼ぶ。キャッシュは無効（常に最新）。
export async function serverGet<T>(
  path: string,
  params?: Record<string, string>
): Promise<T> {
  const cookieStore = await cookies();
  const token = cookieStore.get('access_token')?.value;

  const url = new URL(`${BASE_URL}${path}`);
  if (params) {
    Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
  }

  const res = await fetch(url.toString(), {
    headers: token ? { Cookie: `access_token=${token}` } : {},
    cache: 'no-store',
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}
