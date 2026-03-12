import * as React from "react"
import { ZoomIn } from "lucide-react"
import { cn } from "@/lib/utils"
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog"

interface ScreenshotPreviewProps {
  src: string
  alt: string
  className?: string
}

export function ScreenshotPreview({ src, alt, className }: ScreenshotPreviewProps) {
  const [isOpen, setIsOpen] = React.useState(false)
  const [scale, setScale] = React.useState(1)

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open)
    if (!open) {
      setScale(1)
    }
  }

  const handleZoomIn = () => {
    setScale((prev) => Math.min(prev + 0.5, 3))
  }

  const handleZoomOut = () => {
    setScale((prev) => Math.max(prev - 0.5, 0.5))
  }

  const handleResetZoom = () => {
    setScale(1)
  }

  return (
    <>
      <div
        className={cn(
          "relative inline-block cursor-pointer overflow-hidden rounded-lg group",
          className
        )}
        onClick={() => setIsOpen(true)}
      >
        <img
          src={src}
          alt={alt}
          className="h-[120px] w-auto object-cover rounded-lg transition-transform duration-200"
        />
        <div className="absolute inset-0 flex items-center justify-center bg-black/0 group-hover:bg-black/30 transition-colors duration-200 rounded-lg">
          <ZoomIn className="h-8 w-8 text-white opacity-0 group-hover:opacity-100 transition-opacity duration-200" />
        </div>
      </div>

      <Dialog open={isOpen} onOpenChange={handleOpenChange}>
        <DialogContent className="max-w-[90vw] max-h-[90vh] p-0 overflow-hidden bg-transparent border-0">
          <DialogTitle className="sr-only">{alt}</DialogTitle>
          <div className="relative flex flex-col items-center">
            <div className="absolute top-2 right-2 z-10 flex gap-2">
              <button
                onClick={handleZoomOut}
                className="h-8 w-8 flex items-center justify-center rounded-full bg-black/50 text-white hover:bg-black/70 transition-colors"
                aria-label="缩小"
              >
                -
              </button>
              <button
                onClick={handleResetZoom}
                className="h-8 w-8 flex items-center justify-center rounded-full bg-black/50 text-white hover:bg-black/70 transition-colors text-xs"
                aria-label="重置缩放"
              >
                {Math.round(scale * 100)}%
              </button>
              <button
                onClick={handleZoomIn}
                className="h-8 w-8 flex items-center justify-center rounded-full bg-black/50 text-white hover:bg-black/70 transition-colors"
                aria-label="放大"
              >
                +
              </button>
            </div>
            <div className="overflow-auto max-h-[90vh] max-w-[90vw] flex items-center justify-center p-4">
              <img
                src={src}
                alt={alt}
                style={{ transform: `scale(${scale})`, transformOrigin: 'center' }}
                className="max-w-full max-h-full object-contain transition-transform duration-200"
              />
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
