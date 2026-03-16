import { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import type { TimeRangeStats, TimeRange } from "@/services/statsApi";

interface PieChartCardProps {
  title: string;
  data: Record<TimeRange, TimeRangeStats>;
  className?: string;
  showCenterText?: boolean;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{
    name: string;
    value: number;
    color: string;
  }>;
}

const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
  if (active && payload && payload.length) {
    return (
      <div className="bg-surface border border-outline rounded-lg shadow-level-2 p-3 min-w-[140px]">
        <div className="space-y-1">
          {payload.map((entry, index) => (
            <div key={index} className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <span 
                  className="w-3 h-3 rounded-full" 
                  style={{ backgroundColor: entry.color }}
                />
                <span className="text-sm text-on-surface-variant">{entry.name}</span>
              </div>
              <span className="text-sm font-medium text-on-surface">
                {Number(entry.value).toLocaleString()}
              </span>
            </div>
          ))}
        </div>
      </div>
    );
  }
  return null;
};

interface CustomLegendProps {
  payload?: Array<{
    value: string;
    color: string;
    payload: { value: number };
  }>;
}

const CustomLegend = ({ payload }: CustomLegendProps) => {
  if (payload && payload.length) {
    return (
      <div className="flex items-center justify-center gap-6 mt-2">
        {payload.map((entry, index) => (
          <div key={index} className="flex items-center gap-2">
            <span 
              className="w-2.5 h-2.5 rounded-full" 
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-sm text-on-surface-variant">
              {entry.value}: {entry.payload?.value?.toLocaleString() || 0}
            </span>
          </div>
        ))}
      </div>
    );
  }
  return null;
};

export function PieChartCard({ title, data, className, showCenterText = true }: PieChartCardProps) {
  const { t } = useTranslation();
  const [timeRange, setTimeRange] = useState<TimeRange>("today");
  const stats = data[timeRange];

  const chartColors = useMemo(() => ({
    passed: getComputedStyle(document.documentElement).getPropertyValue("--chart-4").trim() || "#4CAF50",
    failed: getComputedStyle(document.documentElement).getPropertyValue("--chart-5").trim() || "#F44336",
    skipped: getComputedStyle(document.documentElement).getPropertyValue("--chart-3").trim() || "#78909C",
  }), []);

  const chartData = [
    { name: t("dashboard.passed"), value: stats.passed_cases, color: chartColors.passed },
    { name: t("dashboard.failed"), value: stats.failed_cases, color: chartColors.failed },
    { name: t("dashboard.skipped"), value: stats.skipped_cases, color: chartColors.skipped },
  ].filter((item) => item.value > 0);

  const total = stats.passed_cases + stats.failed_cases + stats.skipped_cases;

  const renderCenterText = () => {
    if (!showCenterText || total === 0) return null;
    return (
      <div className="flex flex-col items-center justify-center pointer-events-none">
        <span className="text-3xl font-bold text-foreground">{total.toLocaleString()}</span>
        <span className="text-xs text-muted-foreground mt-0.5">{t("dashboard.totalCases")}</span>
      </div>
    );
  };

  return (
    <Card elevation={1} hoverable={false} className={cn("transition-all duration-200", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        <Select value={timeRange} onValueChange={(v) => setTimeRange(v as TimeRange)}>
          <SelectTrigger className="w-[120px] h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="today">{t("dashboard.today")}</SelectItem>
            <SelectItem value="week">{t("dashboard.thisWeek")}</SelectItem>
            <SelectItem value="month">{t("dashboard.thisMonth")}</SelectItem>
            <SelectItem value="year">{t("dashboard.thisYear")}</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        {total === 0 ? (
          <div className="flex items-center justify-center h-[250px] text-muted-foreground">
            {t("dashboard.noData")}
          </div>
        ) : (
          <div className="h-[250px] relative">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={chartData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={90}
                  paddingAngle={2}
                  dataKey="value"
                >
                  {chartData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip content={<CustomTooltip />} />
                <Legend content={<CustomLegend />} verticalAlign="bottom" height={36} />
              </PieChart>
            </ResponsiveContainer>
            <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 -mt-6">
              {renderCenterText()}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
