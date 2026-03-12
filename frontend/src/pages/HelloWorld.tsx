import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader } from "@/components/ui/card";
import { useNavigate } from "react-router";
import { useState } from "react";
import { CheckCircle, Zap, Layers, Shield, LogOut, Home } from "lucide-react";
import { useTranslation } from "react-i18next";

function HelloWorld() {
  const navigate = useNavigate();
  const [isVisible] = useState(true);
  const { t } = useTranslation();

  const handleLogout = () => {
    localStorage.removeItem("isLoggedIn");
    localStorage.removeItem("user");
    navigate("/login");
  };

  const features = [
    {
      title: t('helloWorld.responsiveDesign'),
      description: t('helloWorld.responsiveDescription'),
      icon: <CheckCircle className="w-6 h-6" />,
      color: "from-blue-500 to-blue-600",
      bg: "bg-blue-50 dark:bg-blue-950/30",
      text: "text-blue-700 dark:text-blue-300",
      subtext: "text-blue-600 dark:text-blue-400"
    },
    {
      title: t('helloWorld.modernTech'),
      description: t('helloWorld.modernTechDescription'),
      icon: <Zap className="w-6 h-6" />,
      color: "from-purple-500 to-purple-600",
      bg: "bg-purple-50 dark:bg-purple-950/30",
      text: "text-purple-700 dark:text-purple-300",
      subtext: "text-purple-600 dark:text-purple-400"
    },
    {
      title: t('helloWorld.componentArchitecture'),
      description: t('helloWorld.componentArchitectureDescription'),
      icon: <Layers className="w-6 h-6" />,
      color: "from-green-500 to-green-600",
      bg: "bg-green-50 dark:bg-green-950/30",
      text: "text-green-700 dark:text-green-300",
      subtext: "text-green-600 dark:text-green-400"
    },
    {
      title: t('helloWorld.security'),
      description: t('helloWorld.securityDescription'),
      icon: <Shield className="w-6 h-6" />,
      color: "from-orange-500 to-orange-600",
      bg: "bg-orange-50 dark:bg-orange-950/30",
      text: "text-orange-700 dark:text-orange-300",
      subtext: "text-orange-600 dark:text-orange-400"
    }
  ];

  return (
    <div className="min-h-screen flex flex-col items-center justify-center py-12 px-4 relative overflow-hidden">
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-96 h-96 bg-gradient-to-br from-blue-400/20 to-indigo-400/20 rounded-full blur-3xl animate-pulse-soft" />
        <div className="absolute -bottom-40 -left-40 w-96 h-96 bg-gradient-to-br from-purple-400/20 to-pink-400/20 rounded-full blur-3xl animate-pulse-soft" style={{ animationDelay: "1.5s" }} />
      </div>

      <div 
        className={`relative z-10 w-full max-w-4xl transition-all duration-700 ease-out ${isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-12'}`}
      >
        <div className="text-center mb-10 animate-slide-up-fade">
          <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-3xl shadow-xl mb-6 animate-float">
            <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h1 className="text-5xl md:text-6xl font-bold bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent mb-4">
            {t('helloWorld.helloTitle')}
          </h1>
          <p className="text-xl text-slate-600 dark:text-slate-400">
            {t('helloWorld.helloSubtitle')}
          </p>
        </div>

        <Card className="backdrop-blur-xl bg-white/80 dark:bg-slate-900/80 border border-white/50 dark:border-slate-700/50 shadow-2xl overflow-hidden">
          <CardHeader className="text-center pb-4 pt-8">
            <CardDescription className="text-lg text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
              {t('helloWorld.description')}
            </CardDescription>
          </CardHeader>
          
          <CardContent className="space-y-8 pb-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-5 px-2">
              {features.map((feature, index) => (
                <div
                  key={index}
                  className={`group relative p-6 rounded-2xl ${feature.bg} border border-transparent hover:border-white/50 dark:hover:border-slate-600/50 transition-all duration-400 hover:shadow-lg hover:-translate-y-1 animate-slide-up-fade`}
                  style={{ animationDelay: `${index * 0.1}s` }}
                >
                  <div className={`absolute inset-0 rounded-2xl bg-gradient-to-br ${feature.color} opacity-0 group-hover:opacity-5 transition-opacity duration-400`} />
                  <div className="relative z-10">
                    <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${feature.color} flex items-center justify-center mb-4 shadow-md group-hover:scale-110 transition-transform duration-300`}>
                      <div className="text-white">
                        {feature.icon}
                      </div>
                    </div>
                    <h3 className={`text-lg font-bold mb-1 ${feature.text}`}>
                      {feature.title}
                    </h3>
                    <p className={`text-sm ${feature.subtext}`}>
                      {feature.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>

            <div className="text-center pt-4">
              <div className="inline-flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 rounded-full border border-blue-100 dark:border-blue-800/30">
                <CheckCircle className="w-4 h-4 text-green-500" />
                <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
                  {t('helloWorld.loginSuccess')}
                </span>
              </div>
            </div>
          </CardContent>
          
          <CardFooter className="flex flex-col sm:flex-row justify-center gap-4 pb-8 pt-4 bg-gradient-to-t from-slate-50/50 to-transparent dark:from-slate-800/20">
            <Button
              variant="secondary"
              onClick={() => navigate("/")}
              className="group w-full sm:w-auto gap-2 px-6 py-6 text-base font-medium bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 transition-all duration-300 shadow-sm hover:shadow"
            >
              <Home className="w-4.5 h-4.5 transition-transform group-hover:scale-110" />
              {t('helloWorld.backToHome')}
            </Button>
            <Button
              variant="default"
              onClick={handleLogout}
              className="group w-full sm:w-auto gap-2 px-6 py-6 text-base font-medium bg-gradient-to-r from-slate-700 to-slate-800 hover:from-slate-800 hover:to-slate-900 text-white transition-all duration-300 shadow-md hover:shadow-lg hover:-translate-y-0.5"
            >
              <LogOut className="w-4.5 h-4.5 transition-transform group-hover:scale-110" />
              {t('helloWorld.logout')}
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}

export default HelloWorld;
