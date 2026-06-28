'use client';

import { Suspense, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import Alert from '@mui/material/Alert';
import Link from '@mui/material/Link';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import { apiClient } from '@/utils/apiClient';
import { useAuth, AuthUser } from '@/context/AuthContext';
import { LoadingButton } from '@/components/elements/buttonBox/LoadingButton';

const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'メールアドレスを入力してください')
    .email('正しいメールアドレスを入力してください'),
  password: z.string().min(1, 'パスワードを入力してください'),
});

type LoginFormValues = z.infer<typeof loginSchema>;
type LoginResponse = { message: string; user: AuthUser };

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const { reload } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);
  const registered = params.get('registered') === '1';

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({ resolver: zodResolver(loginSchema) });

  const onSubmit = async (values: LoginFormValues) => {
    setServerError(null);
    try {
      const res = await apiClient.post<LoginResponse>('/api/auth/login', values);
      await reload();
      router.push(res.user.role === 'admin' ? '/admin' : '/');
    } catch (e) {
      setServerError(e instanceof Error ? e.message : 'ログインに失敗しました');
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'background.default',
        p: 2,
      }}
    >
      <Paper sx={{ p: 4, width: '100%', maxWidth: 400 }} elevation={3}>
        <Typography variant="h5" fontWeight={600} textAlign="center" mb={3}>
          ログイン
        </Typography>

        {registered && (
          <Alert severity="success" sx={{ mb: 2 }}>
            登録が完了しました。ログインしてください。
          </Alert>
        )}
        {serverError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {serverError}
          </Alert>
        )}

        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <TextField
            label="メールアドレス"
            fullWidth
            margin="normal"
            autoComplete="email"
            autoFocus
            error={!!errors.email}
            helperText={errors.email?.message}
            {...register('email')}
          />
          <TextField
            label="パスワード"
            type={showPassword ? 'text' : 'password'}
            fullWidth
            margin="normal"
            autoComplete="current-password"
            error={!!errors.password}
            helperText={errors.password?.message}
            {...register('password')}
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    onClick={() => setShowPassword((v) => !v)}
                    edge="end"
                    aria-label="パスワード表示切替"
                  >
                    {showPassword ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
          />
          <Box mt={3}>
            <LoadingButton
              label="ログイン"
              type="submit"
              fullWidth
              loading={isSubmitting}
            />
          </Box>
        </form>

        <Typography variant="body2" textAlign="center" mt={3}>
          アカウントをお持ちでない方は{' '}
          <Link href="/register" underline="hover">
            新規登録
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={null}>
      <LoginForm />
    </Suspense>
  );
}
