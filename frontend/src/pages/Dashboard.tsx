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
import { statsApi, type TimeRangeStats, type TrendData, type RunningTask, type TimeRange } from "@/services/statsApi";

const mockStats: Record<TimeRange, TimeRangeStats> = {
  today: {
    total_cases: 156,
    total_tasks: 12,
    passed_cases: 142,
    failed_cases: 10,
    skipped_cases: 4,
    pass_rate: 91.0,
  },
  week: {
    total_cases: 892,
    total_tasks: 68,
    passed_cases: 834,
    failed_cases: 45,
    skipped_cases: 13,
    pass_rate: 93.5,
  },
  month: {
    total_cases: 3456,
    total_tasks: 245,
    passed_cases: 3210,
    failed_cases: 198,
    skipped_cases: 48,
    pass_rate: 92.9,
  },
  year: {
    total_cases: 42356,
    total_tasks: 2890,
    passed_cases: 39876,
    failed_cases: 2134,
    skipped_cases: 346,
    pass_rate: 94.1,
  },
  all_time: {
    total_cases: 156789,
    total_tasks: 10234,
    passed_cases: 148234,
    failed_cases: 7456,
    skipped_cases: 1099,
    pass_rate: 94.5,
  },
};

const mockTrendData: TrendData[] = [
  { date: "2024-01-01", total_cases: 120, passed_cases: 110, failed_cases: 8, pass_rate: 91.7 },
  { date: "2024-01-02", total_cases: 135, passed_cases: 128, failed_cases: 5, pass_rate: 94.8 },
  { date: "2024-01-03", total_cases: 98, passed_cases: 92, failed_cases: 4, pass_rate: 93.9 },
  { date: "2024-01-04", total_cases: 156, passed_cases: 145, failed_cases: 9, pass_rate: 92.9 },
  { date: "2024-01-05", total_cases: 142, passed_cases: 138, failed_cases: 3, pass_rate: 97.2 },
  { date: "2024-01-06", total_cases: 89, passed_cases: 85, failed_cases: 2, pass_rate: 95.5 },
  { date: "2024-01-07", total_cases: 167, passed_cases: 158, failed_cases: 7, pass_rate: 94.6 },
];

const mockRunningTasks: RunningTask[] = [
  {
    id: 1,
    task_name: "Smoke Test - Main Branch",
    worker_name: "worker-01",
    plan_key: "SMOKE-MAIN",
    start_time: Date.now() / 1000 - 1800,
    progress: 65,
    estimated_end_time: Date.now() / 1000 + 900,
    passed_tests: 45,
    failed_tests: 2,
    total_tests: 72,
  },
  {
    id: 2,
    task_name: "Regression Suite - Release v2.1",
    worker_name: "worker-02",
    plan_key: "REG-V2.1",
    start_time: Date.now() / 1000 - 3600,
    progress: 32,
    estimated_end_time: Date.now() / 1000 + 7200,
    passed_tests: 128,
    failed_tests: 5,
    total_tests: 415,
  },
  {
    id: 3,
    task_name: "API Integration Tests",
    worker_name: "worker-03",
    plan_key: "API-INT",
    start_time: Date.now() / 1000 - 600,
    progress: 85,
    estimated_end_time: Date.now() / 1000 + 120,
    passed_tests: 34,
    failed_tests: 0,
    total_tests: 40,
  },
];

function Dashboard() {
  const { t } = useTranslation();

  const { data: statsData } = useQuery({
    queryKey: ["dashboardStats"],
    queryFn: () => statsApi.getDashboardStats(),
  });

  const { data: trendData } = useQuery({
    queryKey: ["trendData", "week"],
    queryFn: () => statsApi.getTrendData("week"),
  });

  const { data: runningTasksData } = useQuery({
    queryKey: ["runningTasks"],
    queryFn: () => statsApi.getRunningTasks(),
  });

  const stats = statsData?.data?.stats ?? mockStats;
  const trend = trendData?.data?.trend ?? mockTrendData;
  const runningTasks = runningTasksData?.data?.tasks ?? mockRunningTasks;

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

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <PieChartCard title={t("dashboard.caseDistribution")} data={stats} />
        <LineChartCard title={t("dashboard.trendAnalysis")} data={trend} />
      </div>

      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {t("dashboard.runningTasks")}
        </h2>
        {runningTasks.length === 0 ? (
          <div className="flex items-center justify-center py-12 bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
            <p className="text-slate-500 dark:text-slate-400">
              {t("dashboard.noRunningTasks")}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {runningTasks.map((task) => (
              <RunningTaskCard key={task.id} task={task} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default Dashboard;
