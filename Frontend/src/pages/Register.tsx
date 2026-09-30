import { Link, useNavigate } from "react-router-dom";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { KeyRound, Mail, UserRound } from "lucide-react";

import { Button } from "@/components/ui/button";
import { FieldError, Input, Label } from "@/components/ui/form";
import { useRegister } from "@/features/auth/hooks";
import { AuthLayout } from "@/layouts/AuthLayout";

const schema = z
  .object({
    name: z.string().min(2, "Name must be at least 2 characters."),
    email: z
      .string()
      .min(1, "Email is required.")
      .email("Enter a valid email address."),
    password: z.string().min(8, "Password must be at least 8 characters."),
    confirm: z.string(),
  })
  .refine((values) => values.password === values.confirm, {
    message: "Passwords do not match.",
    path: ["confirm"],
  });

type FormValues = z.infer<typeof schema>;

export function RegisterPage() {
  const navigate = useNavigate();
  const registerMutation = useRegister();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", email: "", password: "", confirm: "" },
  });

  const onSubmit = handleSubmit((values) => {
    registerMutation.mutate(
      { name: values.name, email: values.email, password: values.password },
      { onSuccess: () => navigate("/projects", { replace: true }) },
    );
  });

  return (
    <AuthLayout
      title="Create your account"
      subtitle="Start monitoring conversations in minutes."
      footer={
        <span>
          Already have an account?{" "}
          <Link to="/login" className="text-accent hover:underline">
            Sign in
          </Link>
        </span>
      }
    >
      <form onSubmit={onSubmit} className="space-y-4" noValidate>
        <div>
          <Label htmlFor="name">Full name</Label>
          <div className="relative">
            <UserRound className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
            <Input
              id="name"
              autoComplete="name"
              placeholder="Asha Verma"
              className="pl-8"
              {...register("name")}
            />
          </div>
          <FieldError message={errors.name?.message} />
        </div>
        <div>
          <Label htmlFor="email">Work email</Label>
          <div className="relative">
            <Mail className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="you@organisation.in"
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
              autoComplete="new-password"
              placeholder="At least 8 characters"
              className="pl-8"
              {...register("password")}
            />
          </div>
          <FieldError message={errors.password?.message} />
        </div>
        <div>
          <Label htmlFor="confirm">Confirm password</Label>
          <div className="relative">
            <KeyRound className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
            <Input
              id="confirm"
              type="password"
              autoComplete="new-password"
              placeholder="Repeat your password"
              className="pl-8"
              {...register("confirm")}
            />
          </div>
          <FieldError message={errors.confirm?.message} />
        </div>
        {registerMutation.isError ? (
          <p className="border border-neg/30 bg-neg/10 px-3 py-2 text-xs text-neg">
            {registerMutation.error instanceof Error
              ? registerMutation.error.message
              : "Registration failed. Please try again."}
          </p>
        ) : null}
        <Button
          type="submit"
          className="w-full"
          loading={registerMutation.isPending}
        >
          Create account
        </Button>
      </form>
    </AuthLayout>
  );
}
