import { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  Area,
  ComposedChart,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import type { TrendData, TimeRange } from "@/services/statsApi";

interface LineChartCardProps {
  title: string;
  data: TrendData[];
  className?: string;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{
    name: string;
    value: number;
    color: string;
  }>;
  label?: string;
  formatDate: (date: string) => string;
}

const CustomTooltip = ({ active, payload, label, formatDate }: CustomTooltipProps) => {
  if (active && payload && payload.length && label) {
    return (
      <div className="bg-surface border border-outline rounded-lg shadow-level-2 p-3 min-w-[180px]">
        <p className="text-sm font-medium text-on-surface mb-2">
          {formatDate(label)}
        </p>
        <div className="space-y-1">
          {payload.map((entry, index) => (
            <div key={index} className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <span 
                  className="w-3 h-3 rounded-full" 
                  style={{ backgroundColor: entry.color }}
                />
                <span className="text-sm text-on-surface-variant">
                  {entry.name}
                </span>
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
              {entry.value}
            </span>
          </div>
        ))}
      </div>
    );
  }
  return null;
};

export function LineChartCard({ title, data, className }: LineChartCardProps) {
  const { t } = useTranslation();
  const [timeRange, setTimeRange] = useState<TimeRange>("week");

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    if (timeRange === "today") {
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    }
    if (timeRange === "year") {
      return date.toLocaleDateString([], { month: "short" });
    }
    return date.toLocaleDateString([], { month: "short", day: "numeric" });
  };

  const chartColors = useMemo(() => ({
    total: getComputedStyle(document.documentElement).getPropertyValue("--chart-1").trim() || "#6750A4",
    passed: getComputedStyle(document.documentElement).getPropertyValue("--chart-4").trim() || "#4CAF50",
    failed: getComputedStyle(document.documentElement).getPropertyValue("--chart-5").trim() || "#F44336",
  }), []);

  return (
    <Card 
      elevation={1} 
      hoverable={false}
      className={cn("transition-all duration-200", className)}
    >
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
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
        {data.length === 0 ? (
          <div className="flex items-center justify-center h-[280px] text-muted-foreground">
            {t("dashboard.noData")}
          </div>
        ) : (
          <div className="h-[280px]">
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart
                data={data}
                margin={{
                  top: 10,
                  right: 10,
                  left: -10,
                  bottom: 0,
                }}
              >
                <defs>
                  <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartColors.total} stopOpacity={0.15}/>
                    <stop offset="95%" stopColor={chartColors.total} stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorPassed" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartColors.passed} stopOpacity={0.15}/>
                    <stop offset="95%" stopColor={chartColors.passed} stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid 
                  strokeDasharray="4 4" 
                  stroke="var(--outline)" 
                  opacity={0.4} 
                  vertical={false}
                />
                <XAxis
                  dataKey="date"
                  tickFormatter={formatDate}
                  stroke="var(--on-surface-variant)"
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  dy={10}
                />
                <YAxis 
                  stroke="var(--on-surface-variant)" 
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(value) => value.toLocaleString()}
                />
                <Tooltip 
                  content={<CustomTooltip formatDate={formatDate} />}
                  cursor={{ stroke: "var(--outline)", strokeWidth: 1, strokeDasharray: "4 4" }}
                />
                <Legend content={<CustomLegend />} />
                <Area
                  type="monotone"
                  dataKey="total_cases"
                  name={t("dashboard.totalCases")}
                  stroke="none"
                  fillOpacity={1}
                  fill="url(#colorTotal)"
                />
                <Area
                  type="monotone"
                  dataKey="passed_cases"
                  name={t("dashboard.passedCases")}
                  stroke="none"
                  fillOpacity={1}
                  fill="url(#colorPassed)"
                />
                <Line
                  type="monotone"
                  dataKey="total_cases"
                  name={t("dashboard.totalCases")}
                  stroke={chartColors.total}
                  strokeWidth={2.5}
                  dot={{ r: 3, fill: chartColors.total, strokeWidth: 0 }}
                  activeDot={{ r: 6, fill: chartColors.total, strokeWidth: 0 }}
                />
                <Line
                  type="monotone"
                  dataKey="passed_cases"
                  name={t("dashboard.passedCases")}
                  stroke={chartColors.passed}
                  strokeWidth={2.5}
                  dot={{ r: 3, fill: chartColors.passed, strokeWidth: 0 }}
                  activeDot={{ r: 6, fill: chartColors.passed, strokeWidth: 0 }}
                />
                <Line
                  type="monotone"
                  dataKey="failed_cases"
                  name={t("dashboard.failedCases")}
                  stroke={chartColors.failed}
                  strokeWidth={2.5}
                  dot={{ r: 3, fill: chartColors.failed, strokeWidth: 0 }}
                  activeDot={{ r: 6, fill: chartColors.failed, strokeWidth: 0 }}
                />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
