import { NextRequest, NextResponse } from 'next/server';

const PUBLIC_PATHS = ['/login', '/register'];
const ADMIN_PATHS = ['/admin', '/settings'];

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080';

export async function middleware(request: NextRequest) {
  const token = request.cookies.get('access_token');
  const { pathname } = request.nextUrl;

  const isPublic = PUBLIC_PATHS.includes(pathname);
  const isAdminPath = ADMIN_PATHS.some((p) => pathname.startsWith(p));

  // 未認証 → ログインへ
  if (!token && !isPublic) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // ログイン済みで公開パス → ダッシュボードへ
  if (token && isPublic) {
    return NextResponse.redirect(new URL('/', request.url));
  }

  // admin専用パス（/admin/*・/settings）→ roleをAPIで検証
  if (token && isAdminPath) {
    try {
      const res = await fetch(`${API_BASE}/api/auth/me`, {
        headers: { Cookie: `access_token=${token.value}` },
        cache: 'no-store',
      });
      if (!res.ok) {
        return NextResponse.redirect(new URL('/login', request.url));
      }
      const user = await res.json();
      if (user.role !== 'admin') {
        return NextResponse.redirect(new URL('/', request.url));
      }
    } catch {
      return NextResponse.redirect(new URL('/login', request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
