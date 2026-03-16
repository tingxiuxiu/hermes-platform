import { useNavigate } from "react-router";
import { useTranslation } from "react-i18next";
import { Clock, Play, CheckCircle2, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { RippleEffect } from "@/components/ui/Ripple";
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
        "relative overflow-hidden",
        className
      )}
      elevation={1}
      hoverable={true}
      onClick={handleClick}
    >
      <RippleEffect color="var(--primary)" />
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <CardTitle className="text-base font-medium text-card-foreground truncate">
              {task.task_name}
            </CardTitle>
            <CardDescription className="mt-1 text-sm">
              {task.worker_name} · {task.plan_key}
            </CardDescription>
          </div>
          <Badge variant="filled-tonal" className="flex-shrink-0">
            <Play className="w-3.5 h-3.5" />
            {t("dashboard.running")}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="pt-0">
        <div className="space-y-4">
          <div>
            <div className="flex items-center justify-between text-sm mb-2">
              <span className="text-muted-foreground">{t("dashboard.progress")}</span>
              <span className="font-medium text-primary">{Math.round(progress)}%</span>
            </div>
            <Progress value={progress} className="h-2" />
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4 text-success" />
                <span className="text-sm font-medium text-success">
                  {task.passed_tests}
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <XCircle className="w-4 h-4 text-destructive" />
                <span className="text-sm font-medium text-destructive">
                  {task.failed_tests}
                </span>
              </div>
              <span className="text-sm text-muted-foreground">
                / {task.total_tests}
              </span>
            </div>
            {task.estimated_end_time > 0 && (
              <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <Clock className="w-4 h-4" />
                <span>{t("dashboard.estimatedEnd")}: {formatEstimatedTime(task.estimated_end_time)}</span>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
