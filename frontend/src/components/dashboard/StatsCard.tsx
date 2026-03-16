import React from "react";
import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export interface StatsCardProps {
  title: string;
  value: number | string;
  subtitle?: string;
  trend?: number;
  trendLabel?: string;
  icon?: React.ReactNode;
  iconVariant?: "primary" | "secondary" | "tertiary" | "success" | "warning" | "destructive";
  className?: string;
}

export function StatsCard({
  title,
  value,
  subtitle,
  trend,
  trendLabel,
  icon,
  iconVariant = "primary",
  className,
}: StatsCardProps) {
  const getTrendIcon = () => {
    if (trend === undefined) return null;
    if (trend > 0) {
      return <TrendingUp className="h-3 w-3" />;
    }
    if (trend < 0) {
      return <TrendingDown className="h-3 w-3" />;
    }
    return <Minus className="h-3 w-3" />;
  };

  const getTrendClasses = () => {
    if (trend === undefined) return "bg-muted text-muted-foreground";
    if (trend > 0) return "bg-success-container text-success-container-foreground";
    if (trend < 0) return "bg-destructive-container text-destructive-container-foreground";
    return "bg-muted text-muted-foreground";
  };

  const getIconContainerClasses = () => {
    const variants = {
      primary: "bg-primary-container text-primary-container-foreground",
      secondary: "bg-secondary-container text-secondary-container-foreground",
      tertiary: "bg-tertiary-container text-tertiary-container-foreground",
      success: "bg-success-container text-success-container-foreground",
      warning: "bg-warning-container text-warning-container-foreground",
      destructive: "bg-destructive-container text-destructive-container-foreground",
    };
    return variants[iconVariant];
  };

  return (
    <Card 
      elevation={1} 
      hoverable={false}
      className={cn("transition-all duration-200", className)}
    >
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-3">
        <div className="space-y-1">
          <CardTitle className="text-sm font-medium text-muted-foreground">
            {title}
          </CardTitle>
        </div>
        {icon && (
          <div className={cn(
            "flex items-center justify-center w-10 h-10 rounded-xl transition-colors",
            getIconContainerClasses()
          )}>
            <div className="h-5 w-5">
              {icon}
            </div>
          </div>
        )}
      </CardHeader>
      <CardContent className="space-y-2">
        <div className="text-3xl font-bold tracking-tight text-foreground">
          {typeof value === "number" ? value.toLocaleString() : value}
        </div>
        {subtitle && (
          <p className="text-sm text-muted-foreground">
            {subtitle}
          </p>
        )}
        {trend !== undefined && (
          <div className="flex items-center gap-1 mt-1">
            <div className={cn(
              "flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium",
              getTrendClasses()
            )}>
              {getTrendIcon()}
              <span>
                {trend > 0 ? "+" : ""}
                {trend}%
              </span>
            </div>
            {trendLabel && (
              <span className="text-xs text-muted-foreground">
                {trendLabel}
              </span>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
