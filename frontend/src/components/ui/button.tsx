import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Slot } from "radix-ui"

import { cn } from "@/lib/utils"
import { RippleEffect } from "./Ripple"

const buttonVariants = cva(
  "group/button relative inline-flex shrink-0 items-center justify-center rounded-full border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all duration-200 ease-in-out outline-none select-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-38 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-5",
  {
    variants: {
      variant: {
        filled:
          "bg-primary text-primary-foreground shadow-level-1 hover:shadow-level-2 active:shadow-level-0 hover:bg-primary/92 active:bg-primary/88",
        "filled-tonal":
          "bg-primary-container text-primary-container-foreground shadow-level-1 hover:shadow-level-2 active:shadow-level-0 hover:bg-primary-container/92 active:bg-primary-container/88",
        outlined:
          "border-border bg-background text-primary hover:bg-primary/8 active:bg-primary/12 dark:border-border dark:bg-background dark:hover:bg-primary/8 dark:active:bg-primary/12",
        outline:
          "border-border bg-background text-primary hover:bg-primary/8 active:bg-primary/12 dark:border-border dark:bg-background dark:hover:bg-primary/8 dark:active:bg-primary/12",
        text:
          "text-primary hover:bg-primary/8 active:bg-primary/12",
        secondary:
          "bg-secondary-container text-secondary-container-foreground shadow-level-1 hover:shadow-level-2 active:shadow-level-0 hover:bg-secondary-container/92 active:bg-secondary-container/88",
        ghost:
          "text-foreground hover:bg-muted/50 active:bg-muted dark:hover:bg-muted/50 dark:active:bg-muted",
        destructive:
          "bg-destructive text-destructive-foreground shadow-level-1 hover:shadow-level-2 active:shadow-level-0 hover:bg-destructive/92 active:bg-destructive/88",
        link: "text-primary underline-offset-4 hover:underline",
        default:
          "bg-primary text-primary-foreground shadow-level-1 hover:shadow-level-2 active:shadow-level-0 hover:bg-primary/92 active:bg-primary/88",
      },
      size: {
        default:
          "h-10 gap-2 px-6 has-data-[icon=inline-end]:pr-5 has-data-[icon=inline-start]:pl-5",
        xs: "h-8 gap-1 px-4 text-xs [&_svg:not([class*='size-'])]:size-4",
        sm: "h-9 gap-1.5 px-5 text-[0.875rem] [&_svg:not([class*='size-'])]:size-4.5",
        lg: "h-12 gap-2 px-7 text-base [&_svg:not([class*='size-'])]:size-6",
        icon: "size-10",
        "icon-xs": "size-8 [&_svg:not([class*='size-'])]:size-4",
        "icon-sm": "size-9 [&_svg:not([class*='size-'])]:size-4.5",
        "icon-lg": "size-12 [&_svg:not([class*='size-'])]:size-6",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Button({
  className,
  variant = "default",
  size = "default",
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot.Root : "button"
  
  const getRippleColor = () => {
    if (variant === "filled" || variant === "filled-tonal" || variant === "secondary" || variant === "destructive" || variant === "default") {
      return "white"
    }
    return undefined
  }

  return (
    <Comp
      data-slot="button"
      data-variant={variant}
      data-size={size}
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    >
      {!props.disabled && <RippleEffect color={getRippleColor()} />}
      {props.children}
    </Comp>
  )
}

export { Button, buttonVariants }
