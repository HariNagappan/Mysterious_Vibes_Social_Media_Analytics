import { Link, useLocation, useNavigate } from "react-router-dom";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { KeyRound, Mail } from "lucide-react";

import { DEMO_CREDENTIALS } from "@/app/constants";
import { Button } from "@/components/ui/button";
import { FieldError, Input, Label } from "@/components/ui/form";
import { useLogin } from "@/features/auth/hooks";
import { AuthLayout } from "@/layouts/AuthLayout";

const schema = z.object({
  email: z
    .string()
    .min(1, "Email is required.")
    .email("Enter a valid email address."),
  password: z.string().min(8, "Password must be at least 8 characters."),
});

type FormValues = z.infer<typeof schema>;

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useLogin();
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  const from =
    (location.state as { from?: string } | null)?.from ?? "/projects";

  const onSubmit = handleSubmit((values) => {
    login.mutate(values, {
      onSuccess: () => navigate(from, { replace: true }),
    });
  });

  return (
    <AuthLayout
      title="Sign in to PulseGraph"
      subtitle="Continue to your intelligence workspace."
      footer={
        <span>
          New here?{" "}
          <Link to="/register" className="text-accent hover:underline">
            Create an account
          </Link>
        </span>
      }
    >
      <form onSubmit={onSubmit} className="space-y-4" noValidate>
        <div>
          <Label htmlFor="email">Email</Label>
          <div className="relative">
            <Mail className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="analyst@pulsegraph.dev"
              className="pl-8"
              {...register("email")}
            />
          </div>
          <FieldError message={errors.email?.message} />
        </div>
        <div>
          <Label htmlFor="password">Password</Label>
          <div className="relative">
            <KeyRound className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              placeholder="********"
              className="pl-8"
              {...register("password")}
            />
          </div>
          <FieldError message={errors.password?.message} />
        </div>
        {login.isError ? (
          <p className="border border-neg/30 bg-neg/10 px-3 py-2 text-xs text-neg">
            {login.error instanceof Error
              ? login.error.message
              : "Sign in failed. Please try again."}
          </p>
        ) : null}
        <Button type="submit" className="w-full" loading={login.isPending}>
          Sign in
        </Button>
        <button
          type="button"
          onClick={() => {
            setValue("email", DEMO_CREDENTIALS.email);
            setValue("password", DEMO_CREDENTIALS.password);
          }}
          className="focus-ring w-full rounded-md border border-line py-2 text-xs text-muted transition-colors hover:bg-white/5 hover:text-ink"
        >
          Use demo credentials
        </button>
      </form>
    </AuthLayout>
  );
}
