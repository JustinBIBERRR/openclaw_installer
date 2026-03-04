import { useEffect, useState } from "react";
import { Minus, X } from "lucide-react";

interface Props {
  title: string;
}

const isTauri = typeof window !== "undefined" && "__TAURI__" in window;

export default function TitleBar({ title }: Props) {
  const [win, setWin] = useState<{ minimize: () => void; close: () => void } | null>(null);

  useEffect(() => {
    if (isTauri) {
      import("@tauri-apps/api/window").then(({ getCurrentWindow }) => {
        setWin(getCurrentWindow());
      });
    }
  }, []);

  return (
    <div
      data-tauri-drag-region
      className="h-10 flex items-center justify-between px-4 bg-gray-950 border-b border-gray-800 flex-shrink-0"
    >
      <div className="flex items-center gap-2" data-tauri-drag-region>
        <span className="text-brand-400 font-semibold text-sm">🦞</span>
        <span className="text-gray-300 text-sm font-medium">{title}</span>
      </div>
      {win && (
        <div className="flex items-center gap-1">
          <button
            onClick={() => win.minimize()}
            className="w-7 h-7 flex items-center justify-center rounded hover:bg-gray-700 text-gray-400 hover:text-gray-200 transition-colors"
          >
            <Minus size={13} />
          </button>
          <button
            onClick={() => win.close()}
            className="w-7 h-7 flex items-center justify-center rounded hover:bg-red-600 text-gray-400 hover:text-white transition-colors"
          >
            <X size={13} />
          </button>
        </div>
      )}
    </div>
  );
}
