import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Slot } from "radix-ui"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "group/badge inline-flex h-6 w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-full border border-transparent px-3 py-0.5 text-xs font-medium whitespace-nowrap transition-all duration-200 ease-in-out focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&>svg]:pointer-events-none [&>svg]:size-3!",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-primary/92 [a]:active:bg-primary/88",
        filled: "bg-primary text-primary-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-primary/92 [a]:active:bg-primary/88",
        "filled-tonal": "bg-primary-container text-primary-container-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-primary-container/92 [a]:active:bg-primary-container/88",
        secondary:
          "bg-secondary-container text-secondary-container-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-secondary-container/92 [a]:active:bg-secondary-container/88",
        destructive:
          "bg-destructive-container text-destructive-container-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-destructive-container/92 [a]:active:bg-destructive-container/88",
        success:
          "bg-success-container text-success-container-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-success-container/92 [a]:active:bg-success-container/88",
        warning:
          "bg-warning-container text-warning-container-foreground shadow-level-1 hover:shadow-level-2 [a]:hover:bg-warning-container/92 [a]:active:bg-warning-container/88",
        outline:
          "border-outline text-primary hover:bg-primary/8 active:bg-primary/12 dark:border-outline dark:hover:bg-primary/8 dark:active:bg-primary/12",
        ghost:
          "text-primary hover:bg-primary/8 active:bg-primary/12 dark:hover:bg-primary/8 dark:active:bg-primary/12",
        link: "text-primary underline-offset-4 hover:underline",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Badge({
  className,
  variant = "default",
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot.Root : "span"

  return (
    <Comp
      data-slot="badge"
      data-variant={variant}
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
