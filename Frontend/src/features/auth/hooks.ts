import { useMutation } from "@tanstack/react-query";

import { authAPI } from "@/services/endpoints";

import { useAuthStore } from "./store";

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput extends LoginInput {
  name: string;
}

export function useLogin() {
  const setSession = useAuthStore((state) => state.setSession);
  return useMutation({
    mutationFn: (input: LoginInput) => authAPI.login(input),
    onSuccess: (res) => setSession(res.token, res.user),
  });
}

export function useRegister() {
  const setSession = useAuthStore((state) => state.setSession);
  return useMutation({
    mutationFn: (input: RegisterInput) => authAPI.register(input),
    onSuccess: (res) => setSession(res.token, res.user),
  });
}
