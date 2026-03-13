import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import {
  TestTube,
  ListTodo,
  CheckCircle,
  TrendingUp,
  Clock,
} from "lucide-react";
import { StatsCard } from "@/components/dashboard/StatsCard";
import { PieChartCard } from "@/components/dashboard/PieChartCard";
import { LineChartCard } from "@/components/dashboard/LineChartCard";
import { RunningTaskCard } from "@/components/dashboard/RunningTaskCard";
import { statsApi } from "@/services/statsApi";

function Dashboard() {
  const { t } = useTranslation();

  const { data: statsData, isLoading: statsLoading, error: statsError } = useQuery({
    queryKey: ["dashboardStats"],
    queryFn: () => statsApi.getDashboardStats(),
  });

  const { data: trendData, isLoading: trendLoading, error: trendError } = useQuery({
    queryKey: ["trendData", "week"],
    queryFn: () => statsApi.getTrendData("week"),
  });

  const { data: runningTasksData, isLoading: tasksLoading, error: tasksError } = useQuery({
    queryKey: ["runningTasks"],
    queryFn: () => statsApi.getRunningTasks(),
  });

  const stats = statsData?.data?.stats;
  const trend = trendData?.data?.trend;
  const runningTasks = runningTasksData?.data?.tasks;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
          {t("dashboard.title")}
        </h1>
        <p className="text-sm text-slate-500 dark:text-slate-400">
          {new Date().toLocaleDateString(undefined, {
            weekday: "long",
            year: "numeric",
            month: "long",
            day: "numeric",
          })}
        </p>
      </div>

      {statsLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          {[...Array(5)].map((_, index) => (
            <div key={index} className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 animate-pulse">
              <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-2/3 mb-2"></div>
              <div className="h-8 bg-slate-200 dark:bg-slate-700 rounded w-1/2 mb-2"></div>
              <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-1/3"></div>
            </div>
          ))}
        </div>
      ) : statsError ? (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-red-700 dark:text-red-300">{t("dashboard.loadingError")}</p>
        </div>
      ) : stats ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <StatsCard
            title={t("dashboard.todayCases")}
            value={stats.today.total_cases}
            subtitle={`${t("dashboard.passRate")}: ${stats.today.pass_rate}%`}
            icon={<TestTube className="h-4 w-4" />}
          />
          <StatsCard
            title={t("dashboard.todayTasks")}
            value={stats.today.total_tasks}
            subtitle={`${t("dashboard.passed")}: ${stats.today.passed_cases}`}
            icon={<ListTodo className="h-4 w-4" />}
          />
          <StatsCard
            title={t("dashboard.weekCases")}
            value={stats.week.total_cases}
            subtitle={`${t("dashboard.passRate")}: ${stats.week.pass_rate}%`}
            icon={<TrendingUp className="h-4 w-4" />}
          />
          <StatsCard
            title={t("dashboard.monthCases")}
            value={stats.month.total_cases}
            subtitle={`${t("dashboard.passRate")}: ${stats.month.pass_rate}%`}
            icon={<CheckCircle className="h-4 w-4" />}
          />
          <StatsCard
            title={t("dashboard.allTimeCases")}
            value={stats.all_time.total_cases}
            subtitle={`${t("dashboard.totalTasks")}: ${stats.all_time.total_tasks.toLocaleString()}`}
            icon={<Clock className="h-4 w-4" />}
          />
        </div>
      ) : (
        <div className="bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-4">
          <p className="text-slate-500 dark:text-slate-400">{t("dashboard.noData")}</p>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {trendLoading ? (
          <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 animate-pulse">
            <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-1/3 mb-4"></div>
            <div className="h-64 bg-slate-200 dark:bg-slate-700 rounded"></div>
          </div>
        ) : trendError ? (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <p className="text-red-700 dark:text-red-300">{t("dashboard.loadingError")}</p>
          </div>
        ) : trend ? (
          <LineChartCard title={t("dashboard.trendAnalysis")} data={trend} />
        ) : (
          <div className="bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-4">
            <p className="text-slate-500 dark:text-slate-400">{t("dashboard.noData")}</p>
          </div>
        )}

        {statsLoading ? (
          <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 animate-pulse">
            <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-1/3 mb-4"></div>
            <div className="h-64 bg-slate-200 dark:bg-slate-700 rounded"></div>
          </div>
        ) : statsError ? (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <p className="text-red-700 dark:text-red-300">{t("dashboard.loadingError")}</p>
          </div>
        ) : stats ? (
          <PieChartCard title={t("dashboard.caseDistribution")} data={stats} />
        ) : (
          <div className="bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-4">
            <p className="text-slate-500 dark:text-slate-400">{t("dashboard.noData")}</p>
          </div>
        )}
      </div>

      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {t("dashboard.runningTasks")}
        </h2>
        {tasksLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(3)].map((_, index) => (
              <div key={index} className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 animate-pulse">
                <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-2/3 mb-2"></div>
                <div className="h-3 bg-slate-200 dark:bg-slate-700 rounded w-1/2 mb-3"></div>
                <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded w-full mb-2"></div>
                <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded w-4/5 mb-2"></div>
                <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded w-3/5"></div>
              </div>
            ))}
          </div>
        ) : tasksError ? (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <p className="text-red-700 dark:text-red-300">{t("dashboard.loadingError")}</p>
          </div>
        ) : runningTasks && runningTasks.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {runningTasks.map((task) => (
              <RunningTaskCard key={task.id} task={task} />
            ))}
          </div>
        ) : (
          <div className="flex items-center justify-center py-12 bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
            <p className="text-slate-500 dark:text-slate-400">
              {t("dashboard.noRunningTasks")}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

export default Dashboard;
