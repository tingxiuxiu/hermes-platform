import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useNavigate, useLocation } from "react-router";
import { useTranslation } from "react-i18next";
import { Lock, Mail, Eye, EyeOff, AlertCircle } from "lucide-react";
import { authApi, type LoginRequest } from "@/services/authApi";
import { useAuthStore } from "@/stores/authStore";
import { setToken } from "@/lib/api";
import { encryptPassword } from "@/lib/rsa";

const loginSchema = z.object({
  email: z
    .string()
    .min(1, "login.validation.emailRequired")
    .email("login.validation.emailInvalid"),
  password: z
    .string()
    .min(1, "login.validation.passwordRequired")
    .min(6, "login.validation.passwordMinLength"),
  rememberMe: z.boolean().optional(),
});

type LoginFormData = z.infer<typeof loginSchema>;

function Login() {
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const [isVisible] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((state) => state.login);

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || "/";

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: localStorage.getItem('remembered_email') || "",
      password: "",
      rememberMe: localStorage.getItem('remembered_email') ? true : false,
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setErrorMessage(null);
    
    try {
      let loginData: LoginRequest = {
        email: data.email,
        password: data.password,
      };

      try {
        const pubKeyResponse = await authApi.getRSAPubKey();
        if (pubKeyResponse.success && pubKeyResponse.data?.public_key) {
          const encryptedPwd = await encryptPassword(data.password, pubKeyResponse.data.public_key);
          loginData = {
            email: data.email,
            password: "",
            encrypted_pwd: encryptedPwd,
          };
        }
      } catch (rsaError) {
        console.warn("RSA encryption failed, using plaintext:", rsaError);
      }
      
      const response = await authApi.login(loginData);
      
      if (response.success && response.data) {
        const { token } = response.data;
        
        setToken(token);
        
        // Handle remember me
        if (data.rememberMe) {
          localStorage.setItem('remembered_email', data.email);
        } else {
          localStorage.removeItem('remembered_email');
        }
        
        const profileResponse = await authApi.getProfile();
        
        if (profileResponse.success && profileResponse.data) {
          const { user } = profileResponse.data;
          
          login(token, {
            id: String(user.id),
            email: user.email,
            username: user.name,
            role: user.roles?.[0],
          });
          
          navigate(from, { replace: true });
        } else {
          login(token, {
            id: "",
            email: data.email,
          });
          navigate(from, { replace: true });
        }
      } else {
        setErrorMessage(response.error?.message || t("login.loginFailed"));
      }
    } catch (error) {
      console.error("Login error:", error);
      setErrorMessage(t("login.networkError"));
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 dark:from-slate-950 dark:via-indigo-950 dark:to-slate-950 p-4 overflow-hidden relative">
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-blue-400/20 rounded-full blur-3xl animate-pulse-soft" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-400/20 rounded-full blur-3xl animate-pulse-soft" style={{ animationDelay: "1s" }} />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-indigo-400/10 rounded-full blur-3xl animate-pulse-soft" style={{ animationDelay: "2s" }} />
      </div>

      <div 
        className={`relative z-10 w-full max-w-md transition-all duration-700 ease-out ${isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'}`}
      >
        <div className="mb-8 text-center animate-slide-up-fade">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-2xl shadow-lg mb-4 animate-float">
            <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
            {t('login.platformTitle')}
          </h1>
          <p className="text-muted-foreground mt-2">{t('login.platformSubtitle')}</p>
        </div>

        <Card className="backdrop-blur-xl bg-white/80 dark:bg-slate-900/80 border border-white/50 dark:border-slate-700/50 shadow-2xl hover:shadow-blue-500/20 transition-all duration-500 hover:-translate-y-1">
          <CardHeader className="space-y-1 pb-6">
            <CardTitle className="text-2xl font-bold text-center text-slate-800 dark:text-slate-100">
              {t('login.welcomeBack')}
            </CardTitle>
            <CardDescription className="text-center text-slate-600 dark:text-slate-400">
              {t('login.loginPrompt')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              {errorMessage && (
                <div className="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-600 dark:text-red-400 text-sm">
                  <AlertCircle className="h-4 w-4 flex-shrink-0" />
                  <span>{errorMessage}</span>
                </div>
              )}

              <div className="space-y-2.5">
                <Label htmlFor="email" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                  {t('login.emailUsername')}
                </Label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                    <Mail className="h-4.5 w-4.5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
                  </div>
                  <Input
                    id="email"
                    type="text"
                    placeholder={t('login.emailPlaceholder')}
                    {...register("email")}
                    className={`pl-10 pr-4 py-2.5 bg-white/50 dark:bg-slate-800/50 border-slate-200 dark:border-slate-700 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all duration-300 hover:border-slate-300 dark:hover:border-slate-600 ${
                      errors.email ? "border-red-500 focus:border-red-500 focus:ring-red-500/20" : ""
                    }`}
                  />
                </div>
                {errors.email && (
                  <p className="text-sm text-red-500 mt-1">
                    {t(errors.email.message || "login.validation.emailInvalid")}
                  </p>
                )}
              </div>

              <div className="space-y-2.5">
                <Label htmlFor="password" className="text-sm font-medium text-slate-700 dark:text-slate-300">
                  {t('login.password')}
                </Label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                    <Lock className="h-4.5 w-4.5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
                  </div>
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    placeholder={t('login.passwordPlaceholder')}
                    {...register("password")}
                    className={`pl-10 pr-12 py-2.5 bg-white/50 dark:bg-slate-800/50 border-slate-200 dark:border-slate-700 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all duration-300 hover:border-slate-300 dark:hover:border-slate-600 ${
                      errors.password ? "border-red-500 focus:border-red-500 focus:ring-red-500/20" : ""
                    }`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors focus:outline-none"
                  >
                    {showPassword ? (
                      <EyeOff className="h-4.5 w-4.5" />
                    ) : (
                      <Eye className="h-4.5 w-4.5" />
                    )}
                  </button>
                </div>
                {errors.password && (
                  <p className="text-sm text-red-500 mt-1">
                    {t(errors.password.message || "login.validation.passwordMinLength")}
                  </p>
                )}
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <input
                    id="remember-me"
                    type="checkbox"
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-slate-300 rounded"
                    {...register("rememberMe")}
                  />
                  <label htmlFor="remember-me" className="ml-2 block text-sm text-slate-600 dark:text-slate-400">
                    {t('login.rememberMe')}
                  </label>
                </div>
                <a href="/forgot-password" className="text-sm font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 transition-colors">
                  {t('login.forgotPassword')}
                </a>
              </div>

              <Button
                type="submit"
                className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white py-6 text-base font-semibold shadow-lg hover:shadow-xl transition-all duration-300 transform hover:-translate-y-0.5 active:translate-y-0"
                disabled={isSubmitting}
              >
                {isSubmitting ? (
                  <div className="flex items-center gap-2">
                    <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    <span>{t('login.loggingIn')}</span>
                  </div>
                ) : (
                  t('login.loginButton')
                )}
              </Button>
            </form>
          </CardContent>
          <CardFooter className="flex flex-col space-y-3 pt-2 pb-6">
            <div className="relative w-full">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-200 dark:border-slate-700" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-3 bg-white/80 dark:bg-slate-900/80 text-slate-500 dark:text-slate-400">
                  {t('login.or')}
                </span>
              </div>
            </div>

            <div className="text-sm text-muted-foreground text-center">
              {t('login.noAccount')}
              <a href="/register" className="ml-1 font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 transition-colors hover:underline underline-offset-2">
                {t('login.registerNow')}
              </a>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}

export default Login;
