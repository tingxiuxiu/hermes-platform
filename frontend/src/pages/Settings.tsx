import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { authApi } from "@/services/authApi";
import { LanguageSwitcher } from "@/components/ui/languageSwitcher";
import { Settings as SettingsIcon, Globe, Lock, Bell, Palette } from "lucide-react";

function Settings() {
  const { t } = useTranslation();
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

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
