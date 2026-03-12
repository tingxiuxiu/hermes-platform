import { useNavigate } from "react-router";
import { useTranslation } from "react-i18next";
import { Clock, Play, CheckCircle, XCircle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { RunningTask } from "@/services/statsApi";

interface RunningTaskCardProps {
  task: RunningTask;
  className?: string;
}

function formatEstimatedTime(timestamp: number): string {
  if (!timestamp) return "-";
  const date = new Date(timestamp * 1000);
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

export function RunningTaskCard({ task, className }: RunningTaskCardProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const progress = task.total_tests > 0 
    ? Math.round((task.passed_tests + task.failed_tests) / task.total_tests * 100) 
    : 0;

  const handleClick = () => {
    navigate(`/execution/${task.id}`);
  };

  return (
    <Card
      className={cn(
        "bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 cursor-pointer hover:shadow-md transition-shadow",
        className
      )}
      onClick={handleClick}
    >
      <CardContent className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">
              {task.task_name}
            </h3>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {task.worker_name} · {task.plan_key}
            </p>
          </div>
          <Badge variant="outline" className="bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800">
            <Play className="w-3 h-3 mr-1" />
            {t("dashboard.running")}
          </Badge>
        </div>

        <div className="space-y-3">
          <div>
            <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 mb-1">
              <span>{t("dashboard.progress")}</span>
              <span>{progress}%</span>
            </div>
            <Progress value={progress} className="h-2" />
          </div>

          <div className="flex items-center justify-between text-xs">
            <div className="flex items-center gap-3">
              <span className="flex items-center gap-1 text-green-600 dark:text-green-400">
                <CheckCircle className="w-3.5 h-3.5" />
                {task.passed_tests}
              </span>
              <span className="flex items-center gap-1 text-red-600 dark:text-red-400">
                <XCircle className="w-3.5 h-3.5" />
                {task.failed_tests}
              </span>
              <span className="text-slate-500 dark:text-slate-400">
                / {task.total_tests}
              </span>
            </div>
            {task.estimated_end_time > 0 && (
              <span className="flex items-center gap-1 text-slate-500 dark:text-slate-400">
                <Clock className="w-3.5 h-3.5" />
                {t("dashboard.estimatedEnd")}: {formatEstimatedTime(task.estimated_end_time)}
              </span>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
