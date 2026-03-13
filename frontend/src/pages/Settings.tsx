import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { authApi } from "@/services/authApi";
import { tokenApi, type APIToken } from "@/services/tokenApi";
import { LanguageSwitcher } from "@/components/ui/languageSwitcher";
import { Settings as SettingsIcon, Globe, Lock, Bell, Palette, Key, Copy, Trash2, CheckCircle } from "lucide-react";
import { useQuery, useMutation } from "@tanstack/react-query";

function Settings() {
  const { t } = useTranslation();
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const [newTokenName, setNewTokenName] = useState("");
  const [generatedToken, setGeneratedToken] = useState<string | null>(null);

  const { data: tokensData, isLoading: tokensLoading, refetch: refetchTokens } = useQuery({
    queryKey: ["apiTokens"],
    queryFn: () => tokenApi.getTokens(),
  });

  const createTokenMutation = useMutation({
    mutationFn: (name: string) => tokenApi.createToken(name),
    onSuccess: (data) => {
      if (data.success && data.data) {
        setGeneratedToken(data.data.token);
        setNewTokenName("");
        refetchTokens();
      } else {
        setMessage({ type: "error", text: data.error?.message || t("settings.tokenCreateFailed") });
      }
    },
    onError: () => {
      setMessage({ type: "error", text: t("settings.tokenCreateFailed") });
    },
  });

  const revokeTokenMutation = useMutation({
    mutationFn: (id: number) => tokenApi.revokeToken(id),
    onSuccess: (data) => {
      if (data.success) {
        setMessage({ type: "success", text: t("settings.tokenRevoked") });
        refetchTokens();
      } else {
        setMessage({ type: "error", text: data.error?.message || t("settings.tokenRevokeFailed") });
      }
    },
    onError: () => {
      setMessage({ type: "error", text: t("settings.tokenRevokeFailed") });
    },
  });

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage(null);

    if (newPassword !== confirmPassword) {
      setMessage({ type: "error", text: t("settings.passwordMismatch", "新密码与确认密码不匹配") });
      return;
    }

    if (newPassword.length < 6) {
      setMessage({ type: "error", text: t("settings.passwordTooShort", "密码长度至少为6位") });
      return;
    }

    setLoading(true);
    try {
      const response = await authApi.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });

      if (response.success) {
        setMessage({ type: "success", text: t("settings.passwordChanged", "密码修改成功") });
        setOldPassword("");
        setNewPassword("");
        setConfirmPassword("");
      } else {
        setMessage({ type: "error", text: response.error?.message || t("settings.passwordChangeFailed", "密码修改失败") });
      }
    } catch (error) {
      setMessage({ type: "error", text: t("settings.networkError", "网络错误，请稍后重试") });
    } finally {
      setLoading(false);
    }
  };

  const handleCreateToken = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTokenName.trim()) return;
    createTokenMutation.mutate(newTokenName);
  };

  const handleRevokeToken = (id: number) => {
    if (window.confirm(t("settings.confirmRevokeToken", "确定要撤销此 Token 吗？"))) {
      revokeTokenMutation.mutate(id);
    }
  };

  const copyToken = (token: string) => {
    navigator.clipboard.writeText(token);
    setMessage({ type: "success", text: t("settings.tokenCopied", "Token 已复制到剪贴板") });
  };

  const tokens: APIToken[] = tokensData?.data?.tokens || [];

  return (
    <div className="container mx-auto py-6 px-4 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
          <SettingsIcon className="h-6 w-6" />
          {t("settings.title", "设置")}
        </h1>
        <p className="text-slate-600 dark:text-slate-400 mt-1">
          {t("settings.subtitle", "管理您的账户设置和偏好")}
        </p>
      </div>

      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              {t("settings.language", "语言设置")}
            </CardTitle>
            <CardDescription>
              {t("settings.languageDesc", "选择您的首选语言")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4">
              <LanguageSwitcher />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Lock className="h-5 w-5" />
              {t("settings.security", "安全设置")}
            </CardTitle>
            <CardDescription>
              {t("settings.securityDesc", "修改您的登录密码")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleChangePassword} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="old-password">{t("settings.oldPassword", "当前密码")}</Label>
                <Input
                  id="old-password"
                  type="password"
                  value={oldPassword}
                  onChange={(e) => setOldPassword(e.target.value)}
                  placeholder={t("settings.oldPasswordPlaceholder", "请输入当前密码")}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="new-password">{t("settings.newPassword", "新密码")}</Label>
                <Input
                  id="new-password"
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder={t("settings.newPasswordPlaceholder", "请输入新密码")}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm-password">{t("settings.confirmPassword", "确认新密码")}</Label>
                <Input
                  id="confirm-password"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder={t("settings.confirmPasswordPlaceholder", "请再次输入新密码")}
                  required
                />
              </div>
              {message && (
                <div
                  className={`p-3 rounded-lg text-sm ${
                    message.type === "success"
                      ? "bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400"
                      : "bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400"
                  }`}
                >
                  {message.text}
                </div>
              )}
              <Button type="submit" disabled={loading}>
                {loading ? t("settings.saving", "保存中...") : t("settings.savePassword", "保存密码")}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              {t("settings.apiTokens", "API Token")}
            </CardTitle>
            <CardDescription>
              {t("settings.apiTokensDesc", "管理用于自动化测试的 API Token，有效期 365 天")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <form onSubmit={handleCreateToken} className="flex gap-2">
              <Input
                type="text"
                value={newTokenName}
                onChange={(e) => setNewTokenName(e.target.value)}
                placeholder={t("settings.tokenNamePlaceholder", "输入 Token 名称")}
                className="flex-1"
              />
              <Button type="submit" disabled={createTokenMutation.isPending || !newTokenName.trim()}>
                {createTokenMutation.isPending ? t("settings.generating", "生成中...") : t("settings.generateToken", "生成 Token")}
              </Button>
            </form>

            {generatedToken && (
              <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
                <div className="flex items-center gap-2 text-green-600 dark:text-green-400 mb-2">
                  <CheckCircle className="h-4 w-4" />
                  <span className="font-medium">{t("settings.tokenGenerated", "Token 生成成功！")}</span>
                </div>
                <div className="flex items-center gap-2">
                  <code className="flex-1 p-2 bg-white dark:bg-slate-800 rounded text-xs break-all">
                    {generatedToken}
                  </code>
                  <Button variant="outline" size="sm" onClick={() => copyToken(generatedToken)}>
                    <Copy className="h-4 w-4" />
                  </Button>
                </div>
                <p className="text-xs text-green-600 dark:text-green-400 mt-2">
                  {t("settings.tokenWarning", "请立即复制 Token，此 Token 只会显示一次！")}
                </p>
                <Button variant="ghost" size="sm" className="mt-2" onClick={() => setGeneratedToken(null)}>
                  {t("settings.dismiss", "知道了")}
                </Button>
              </div>
            )}

            {tokensLoading ? (
              <div className="text-center py-4 text-slate-500">{t("settings.loading", "加载中...")}</div>
            ) : tokens.length > 0 ? (
              <div className="space-y-2">
                <h4 className="text-sm font-medium">{t("settings.yourTokens", "您的 Token 列表")}</h4>
                {tokens.map((token) => (
                  <div
                    key={token.id}
                    className="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800 rounded-lg"
                  >
                    <div>
                      <div className="font-medium">{token.name}</div>
                      <div className="text-xs text-slate-500">
                        {token.token} · {t("settings.expiresAt", "过期于")} {token.expires_at}
                        {token.is_revoked && ` · ${t("settings.revoked", "已撤销")}`}
                      </div>
                    </div>
                    {!token.is_revoked && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRevokeToken(token.id)}
                        className="text-red-500 hover:text-red-600"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-slate-500 dark:text-slate-400 text-center py-4">
                {t("settings.noTokens", "暂无 API Token")}
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              {t("settings.notifications", "通知设置")}
            </CardTitle>
            <CardDescription>
              {t("settings.notificationsDesc", "管理您的通知偏好")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-slate-500 dark:text-slate-400">
              {t("settings.notificationsComingSoon", "通知功能即将推出...")}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Palette className="h-5 w-5" />
              {t("settings.appearance", "外观设置")}
            </CardTitle>
            <CardDescription>
              {t("settings.appearanceDesc", "自定义应用外观")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-slate-500 dark:text-slate-400">
              {t("settings.themeComingSoon", "主题切换功能即将推出...")}
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default Settings;
