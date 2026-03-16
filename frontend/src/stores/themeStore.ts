import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

export type Theme = "light" | "dark";

export interface ThemeState {
  theme: Theme;
  isDark: boolean;
}

export interface ThemeActions {
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

export type ThemeStore = ThemeState & ThemeActions;

const updateHtmlClass = (theme: Theme) => {
  const html = document.documentElement;
  if (theme === "dark") {
    html.classList.add("dark");
  } else {
    html.classList.remove("dark");
  }
};

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set, get) => {
      const initialTheme = "light";
      updateHtmlClass(initialTheme);

      return {
        theme: initialTheme,
        isDark: false,

        setTheme: (theme: Theme) => {
          updateHtmlClass(theme);
          set({
            theme,
            isDark: theme === "dark",
          });
        },

        toggleTheme: () => {
          const currentTheme = get().theme;
          const newTheme = currentTheme === "light" ? "dark" : "light";
          updateHtmlClass(newTheme);
          set({
            theme: newTheme,
            isDark: newTheme === "dark",
          });
        },
      };
    },
    {
      name: "theme-storage",
      storage: createJSONStorage(() => localStorage),
      onRehydrateStorage: () => (state) => {
        if (state) {
          updateHtmlClass(state.theme);
        }
      },
    },
  ),
);
