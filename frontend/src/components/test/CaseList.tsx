import { useTranslation } from "react-i18next"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import type { TestDetail } from "@/services/testDetailApi"

export interface CaseListProps {
  cases: TestDetail[]
  selectedId: number | null
  onSelect: (testCase: TestDetail) => void
}

function formatDuration(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  return `${hours.toString().padStart(2, "0")}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`
}

function getStatusConfig(status: TestDetail["test_status"]) {
  switch (status) {
    case "passed":
      return {
        bgClass: "bg-green-50 dark:bg-green-900/20",
        badgeVariant: "default" as const,
        badgeClass: "bg-green-500 text-white",
      }
    case "failed":
      return {
        bgClass: "bg-red-50 dark:bg-red-900/20",
        badgeVariant: "destructive" as const,
        badgeClass: "",
      }
    case "skipped":
      return {
        bgClass: "bg-slate-50 dark:bg-slate-800",
        badgeVariant: "secondary" as const,
        badgeClass: "",
      }
  }
}

export function CaseList({ cases, selectedId, onSelect }: CaseListProps) {
  const { t } = useTranslation()

  const getStatusText = (status: TestDetail["test_status"]) => {
    switch (status) {
      case "passed":
        return t("caseList.status.passed", "通过")
      case "failed":
        return t("caseList.status.failed", "失败")
      case "skipped":
        return t("caseList.status.skipped", "跳过")
    }
  }

  return (
    <ScrollArea className="h-full w-full">
      <div className="flex flex-col gap-1 p-2">
        {cases.map((testCase) => {
          const statusConfig = getStatusConfig(testCase.test_status)
          const isSelected = testCase.id === selectedId

          return (
            <div
              key={testCase.id}
              onClick={() => onSelect(testCase)}
              className={cn(
                "flex h-[60px] cursor-pointer items-center justify-between rounded-lg border-2 px-4 transition-all",
                statusConfig.bgClass,
                isSelected
                  ? "border-blue-500 bg-blue-50/50 dark:bg-blue-900/30"
                  : "border-transparent hover:border-blue-300"
              )}
            >
              <div className="flex min-w-0 flex-1 items-center gap-4">
                <span className="truncate font-medium text-slate-700 dark:text-slate-200">
                  {testCase.test_name}
                </span>
                <Badge
                  variant={statusConfig.badgeVariant}
                  className={statusConfig.badgeClass}
                >
                  {getStatusText(testCase.test_status)}
                </Badge>
              </div>
              <div className="ml-4 flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
                <span>{formatDuration(testCase.duration)}</span>
              </div>
            </div>
          )
        })}
        {cases.length === 0 && (
          <div className="flex h-[200px] items-center justify-center text-slate-400">
            {t("caseList.noCases", "暂无用例数据")}
          </div>
        )}
      </div>
    </ScrollArea>
  )
}
