'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Alert from '@mui/material/Alert';
import Link from '@mui/material/Link';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import { apiClient } from '@/utils/apiClient';
import { LoadingButton } from '@/components/elements/buttonBox/LoadingButton';

const passwordToggle = (show: boolean, onToggle: () => void) => ({
  endAdornment: (
    <InputAdornment position="end">
      <IconButton onClick={onToggle} edge="end" aria-label="パスワード表示切替">
        {show ? <VisibilityOff /> : <Visibility />}
      </IconButton>
    </InputAdornment>
  ),
});

const registerSchema = z
  .object({
    invitationCode: z.string().min(1, '招待コードを入力してください'),
    name: z
      .string()
      .min(1, 'お名前を入力してください')
      .max(50, 'お名前は50文字以内で入力してください'),
    email: z
      .string()
      .min(1, 'メールアドレスを入力してください')
      .email('正しいメールアドレスを入力してください'),
    password: z
      .string()
      .min(8, 'パスワードは8文字以上で入力してください')
      .regex(/^(?=.*[a-zA-Z])(?=.*[0-9])/, 'パスワードは英字と数字を含めてください'),
    passwordConfirm: z.string().min(1, 'パスワード（確認）を入力してください'),
  })
  .refine((d) => d.password === d.passwordConfirm, {
    message: 'パスワードが一致しません',
    path: ['passwordConfirm'],
  });

type RegisterFormValues = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({ resolver: zodResolver(registerSchema) });

  const onSubmit = async (values: RegisterFormValues) => {
    setServerError(null);
    try {
      await apiClient.post('/api/auth/register', {
        invitation_code: values.invitationCode,
        name: values.name,
        email: values.email,
        password: values.password,
      });
      router.push('/login?registered=1');
    } catch (e) {
      setServerError(e instanceof Error ? e.message : '登録に失敗しました');
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
      <Paper sx={{ p: 4, width: '100%', maxWidth: 440 }} elevation={3}>
        <Typography variant="h5" fontWeight={600} textAlign="center" mb={3}>
          新規登録
        </Typography>

        {serverError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {serverError}
          </Alert>
        )}

        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <TextField
            label="招待コード"
            fullWidth
            margin="normal"
            autoFocus
            placeholder="TRADE-XXXX-XXXX"
            error={!!errors.invitationCode}
            helperText={errors.invitationCode?.message}
            {...register('invitationCode', {
              onChange: (e) =>
                setValue('invitationCode', e.target.value.toUpperCase()),
            })}
          />
          <TextField
            label="お名前"
            fullWidth
            margin="normal"
            error={!!errors.name}
            helperText={errors.name?.message}
            {...register('name')}
          />
          <TextField
            label="メールアドレス"
            fullWidth
            margin="normal"
            autoComplete="email"
            error={!!errors.email}
            helperText={errors.email?.message}
            {...register('email')}
          />
          <TextField
            label="パスワード"
            type={showPassword ? 'text' : 'password'}
            fullWidth
            margin="normal"
            autoComplete="new-password"
            error={!!errors.password}
            helperText={errors.password?.message}
            {...register('password')}
            InputProps={passwordToggle(showPassword, () => setShowPassword((v) => !v))}
          />
          <TextField
            label="パスワード（確認）"
            type={showConfirm ? 'text' : 'password'}
            fullWidth
            margin="normal"
            autoComplete="new-password"
            error={!!errors.passwordConfirm}
            helperText={errors.passwordConfirm?.message}
            {...register('passwordConfirm')}
            InputProps={passwordToggle(showConfirm, () => setShowConfirm((v) => !v))}
          />
          <Box mt={3}>
            <LoadingButton
              label="登録する"
              type="submit"
              fullWidth
              loading={isSubmitting}
            />
          </Box>
        </form>

        <Typography variant="body2" textAlign="center" mt={3}>
          すでにアカウントをお持ちの方は{' '}
          <Link href="/login" underline="hover">
            ログイン
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
}
