import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import {
  TestTube,
  ListTodo,
  CheckCircle,
  TrendingUp,
  Clock,
  Filter,
  Calendar,
  RefreshCw,
} from "lucide-react";
import { StatsCard } from "@/components/dashboard/StatsCard";
import { PieChartCard } from "@/components/dashboard/PieChartCard";
import { LineChartCard } from "@/components/dashboard/LineChartCard";
import { RunningTaskCard } from "@/components/dashboard/RunningTaskCard";
import { Button } from "@/components/ui/button";
import { statsApi } from "@/services/statsApi";

const SkeletonStats = ({ count = 5 }: { count?: number }) => (
  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
    {[...Array(count)].map((_, index) => (
      <div key={index} className="bg-card rounded-xl border border-border p-6 animate-pulse">
        <div className="flex items-start justify-between">
          <div className="space-y-3">
            <div className="h-4 bg-muted rounded w-24"></div>
            <div className="h-8 bg-muted rounded w-20"></div>
            <div className="h-3 bg-muted rounded w-32"></div>
          </div>
          <div className="h-10 w-10 bg-muted rounded-xl"></div>
        </div>
      </div>
    ))}
  </div>
);

const SkeletonChart = () => (
  <div className="bg-card rounded-xl border border-border p-6 animate-pulse">
    <div className="flex items-center justify-between mb-4">
      <div className="h-4 bg-muted rounded w-32"></div>
      <div className="h-8 bg-muted rounded w-28"></div>
    </div>
    <div className="h-64 bg-muted/50 rounded-lg"></div>
  </div>
);

const SkeletonTasks = () => (
  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    {[...Array(3)].map((_, index) => (
      <div key={index} className="bg-card rounded-xl border border-border p-5 animate-pulse">
        <div className="space-y-3">
          <div className="h-4 bg-muted rounded w-3/4"></div>
          <div className="h-3 bg-muted rounded w-1/2"></div>
          <div className="h-2 bg-muted rounded w-full"></div>
          <div className="h-2 bg-muted rounded w-4/5"></div>
        </div>
      </div>
    ))}
  </div>
);

function Dashboard() {
  const { t } = useTranslation();

  const { data: statsData, isLoading: statsLoading, error: statsError, refetch: refetchStats } = useQuery({
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
    <div className="flex flex-col gap-6 pb-8">
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            {t("dashboard.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("dashboard.subtitle")}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button 
            variant="outlined" 
            size="sm" 
            onClick={() => refetchStats()}
            className="gap-1.5"
          >
            <RefreshCw className="h-4 w-4" />
            {t("dashboard.refresh")}
          </Button>
          <Button variant="outlined" size="sm" className="gap-1.5">
            <Calendar className="h-4 w-4" />
            {t("dashboard.filter")}
          </Button>
          <Button size="sm" className="gap-1.5">
            <Filter className="h-4 w-4" />
            {t("dashboard.actions")}
          </Button>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {new Date().toLocaleDateString(undefined, {
            weekday: "long",
            year: "numeric",
            month: "long",
            day: "numeric",
          })}
        </p>
      </div>

      <section aria-labelledby="stats-heading" className="space-y-4">
        <div className="flex items-center gap-2">
          <h2 id="stats-heading" className="text-sm font-medium text-muted-foreground">
            {t("dashboard.overview")}
          </h2>
        </div>
        {statsLoading ? (
          <SkeletonStats />
        ) : statsError ? (
          <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-6">
            <p className="text-destructive">{t("dashboard.loadingError")}</p>
          </div>
        ) : stats ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
            <StatsCard
              title={t("dashboard.todayCases")}
              value={stats.today.total_cases}
              subtitle={`${t("dashboard.passRate")}: ${stats.today.pass_rate}%`}
              icon={<TestTube className="h-4 w-4" />}
              iconVariant="primary"
            />
            <StatsCard
              title={t("dashboard.todayTasks")}
              value={stats.today.total_tasks}
              subtitle={`${t("dashboard.passed")}: ${stats.today.passed_cases}`}
              icon={<ListTodo className="h-4 w-4" />}
              iconVariant="secondary"
            />
            <StatsCard
              title={t("dashboard.weekCases")}
              value={stats.week.total_cases}
              subtitle={`${t("dashboard.passRate")}: ${stats.week.pass_rate}%`}
              icon={<TrendingUp className="h-4 w-4" />}
              iconVariant="tertiary"
            />
            <StatsCard
              title={t("dashboard.monthCases")}
              value={stats.month.total_cases}
              subtitle={`${t("dashboard.passRate")}: ${stats.month.pass_rate}%`}
              icon={<CheckCircle className="h-4 w-4" />}
              iconVariant="success"
            />
            <StatsCard
              title={t("dashboard.allTimeCases")}
              value={stats.all_time.total_cases}
              subtitle={`${t("dashboard.totalTasks")}: ${stats.all_time.total_tasks.toLocaleString()}`}
              icon={<Clock className="h-4 w-4" />}
              iconVariant="warning"
            />
          </div>
        ) : (
          <div className="bg-muted/30 border border-border rounded-lg p-6">
            <p className="text-muted-foreground">{t("dashboard.noData")}</p>
          </div>
        )}
      </section>

      <section aria-labelledby="charts-heading" className="space-y-4">
        <div className="flex items-center gap-2">
          <h2 id="charts-heading" className="text-sm font-medium text-muted-foreground">
            {t("dashboard.analytics")}
          </h2>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <div className="lg:col-span-2">
            {trendLoading ? (
              <SkeletonChart />
            ) : trendError ? (
                <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-6">
                  <p className="text-destructive">{t("dashboard.loadingError")}</p>
                </div>
              ) : trend ? (
                <LineChartCard title={t("dashboard.trendAnalysis")} data={trend} />
              ) : (
                <div className="bg-muted/30 border border-border rounded-lg p-6">
                  <p className="text-muted-foreground">{t("dashboard.noData")}</p>
                </div>
              )}
          </div>
          <div>
            {statsLoading ? (
              <SkeletonChart />
            ) : statsError ? (
                <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-6">
                  <p className="text-destructive">{t("dashboard.loadingError")}</p>
                </div>
              ) : stats ? (
                <PieChartCard title={t("dashboard.caseDistribution")} data={stats} />
              ) : (
                <div className="bg-muted/30 border border-border rounded-lg p-6">
                  <p className="text-muted-foreground">{t("dashboard.noData")}</p>
                </div>
              )}
          </div>
        </div>
      </section>

      <section aria-labelledby="tasks-heading" className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h2 id="tasks-heading" className="text-sm font-medium text-muted-foreground">
              {t("dashboard.runningTasks")}
            </h2>
          </div>
          {runningTasks && runningTasks.length > 0 && (
            <Button variant="text" size="sm">
            {t("dashboard.viewAll")}
          </Button>
          )}
        </div>
        {tasksLoading ? (
          <SkeletonTasks />
        ) : tasksError ? (
          <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-6">
            <p className="text-destructive">{t("dashboard.loadingError")}</p>
          </div>
        ) : runningTasks && runningTasks.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {runningTasks.map((task) => (
              <RunningTaskCard key={task.id} task={task} />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-16 bg-card rounded-xl border border-border">
            <ListTodo className="h-12 w-12 text-muted-foreground/50 mb-3" />
            <p className="text-muted-foreground font-medium">
              {t("dashboard.noRunningTasks")}
            </p>
            <p className="text-sm text-muted-foreground mt-1">
              {t("dashboard.noRunningTasksDesc")}
            </p>
          </div>
        )}
      </section>
    </div>
  );
}

export default Dashboard;
