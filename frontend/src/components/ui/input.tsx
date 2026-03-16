import * as React from "react"

import { cn } from "@/lib/utils"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      type={type}
      data-slot="input"
      className={cn(
        "h-12 w-full min-w-0 rounded-2xl border border-outline bg-surface-container px-4 py-2 text-base text-on-surface transition-all duration-200 ease-in-out outline-none file:inline-flex file:h-8 file:border-0 file:bg-primary-container file:text-primary-container-foreground file:text-sm file:font-medium file:px-4 file:rounded-full placeholder:text-on-surface-variant hover:border-outline focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-surface-container/50 disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:bg-surface-container dark:disabled:bg-surface-container/80 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40",
        className
      )}
      {...props}
    />
  )
}

export { Input }
