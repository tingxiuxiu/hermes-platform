import * as React from "react"

interface RippleProps {
  color?: string
}

interface Ripple {
  id: number
  x: number
  y: number
  size: number
}

export function RippleEffect({ color }: RippleProps) {
  const [ripples, setRipples] = React.useState<Ripple[]>([])
  const containerRef = React.useRef<HTMLSpanElement>(null)
  const nextIdRef = React.useRef(0)

  const addRipple = React.useCallback(
    (event: React.MouseEvent<HTMLSpanElement>) => {
      const container = containerRef.current
      if (!container) return

      const rect = container.getBoundingClientRect()
      const size = Math.max(rect.width, rect.height)
      const x = event.clientX - rect.left - size / 2
      const y = event.clientY - rect.top - size / 2

      const newRipple: Ripple = {
        id: nextIdRef.current++,
        x,
        y,
        size,
      }

      setRipples((prev) => [...prev, newRipple])

      setTimeout(() => {
        setRipples((prev) => prev.filter((r) => r.id !== newRipple.id))
      }, 600)
    },
    []
  )

  return (
    <span
      ref={containerRef}
      onClick={addRipple}
      className="absolute inset-0 overflow-hidden"
      style={{ color }}
    >
      {ripples.map((ripple) => (
        <span
          key={ripple.id}
          className="ripple"
          style={{
            left: ripple.x,
            top: ripple.y,
            width: ripple.size,
            height: ripple.size,
          }}
        />
      ))}
    </span>
  )
}
