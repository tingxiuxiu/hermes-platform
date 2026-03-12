import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useAuthStore } from "@/stores/authStore";
import { authApi } from "@/services/authApi";
import { User as UserIcon, Mail, Shield, Calendar } from "lucide-react";

function Profile() {
  const { t } = useTranslation();
  const { user, setUser } = useAuthStore();
  const [name, setName] = useState(user?.username || "");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await authApi.getProfile();
        if (response.success && response.data) {
          setUser({
            id: String(response.data.user.id),
            email: response.data.user.email,
            username: response.data.user.name,
            role: response.data.user.roles?.[0],
          });
          setName(response.data.user.name);
        }
      } catch (error) {
        console.error("Failed to fetch profile:", error);
      }
    };

    fetchProfile();
  }, [setUser]);

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage(null);
    setLoading(true);

    try {
      const response = await authApi.updateProfile({ name });
      if (response.success) {
        setMessage({ type: "success", text: t("profile.updateSuccess", "个人资料更新成功") });
        if (user) {
          setUser({ ...user, username: name });
        }
      } else {
        setMessage({ type: "error", text: response.error?.message || t("profile.updateFailed", "更新失败") });
      }
    } catch (error) {
      setMessage({ type: "error", text: t("profile.networkError", "网络错误，请稍后重试") });
    } finally {
      setLoading(false);
    }
  };

  const initials = user?.email?.slice(0, 2).toUpperCase() || "U";

  return (
    <div className="container mx-auto py-6 px-4 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
          <UserIcon className="h-6 w-6" />
          {t("profile.title", "个人资料")}
        </h1>
        <p className="text-slate-600 dark:text-slate-400 mt-1">
          {t("profile.subtitle", "查看和管理您的个人信息")}
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card className="md:col-span-1">
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              <Avatar className="h-24 w-24">
                <AvatarFallback className="bg-blue-500 text-white text-2xl font-medium">
                  {initials}
                </AvatarFallback>
              </Avatar>
            </div>
            <CardTitle>{user?.username || user?.email}</CardTitle>
            <CardDescription>{user?.email}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400">
              <Shield className="h-4 w-4" />
              <span>{t("profile.role", "角色")}: {user?.role || "-"}</span>
            </div>
            <div className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400">
              <Calendar className="h-4 w-4" />
              <span>{t("profile.memberSince", "注册时间")}: -</span>
            </div>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>{t("profile.editProfile", "编辑资料")}</CardTitle>
            <CardDescription>
              {t("profile.editProfileDesc", "更新您的个人信息")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleUpdateProfile} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email" className="flex items-center gap-2">
                  <Mail className="h-4 w-4" />
                  {t("profile.email", "邮箱")}
                </Label>
                <Input
                  id="email"
                  type="email"
                  value={user?.email || ""}
                  disabled
                  className="bg-slate-50 dark:bg-slate-800"
                />
                <p className="text-xs text-slate-500">
                  {t("profile.emailNotChangeable", "邮箱地址不可修改")}
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="name" className="flex items-center gap-2">
                  <UserIcon className="h-4 w-4" />
                  {t("profile.displayName", "显示名称")}
                </Label>
                <Input
                  id="name"
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t("profile.namePlaceholder", "请输入显示名称")}
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
                {loading ? t("profile.saving", "保存中...") : t("profile.saveChanges", "保存更改")}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default Profile;
