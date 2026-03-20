import { Moon, Sun, Bell, Search, Menu } from 'lucide-react'
import { useTheme } from '@/contexts'
import { useState } from 'react'
import { useSidebar } from './Layout'

const HEADER_CLASSES =
  'sticky top-0 z-30 h-16 bg-white/80 dark:bg-gray-900/80 backdrop-blur-xl border-b border-gray-200 dark:border-gray-800'

const HEADER_CONTAINER_CLASSES =
  'h-full px-4 sm:px-6 flex items-center justify-between gap-3 sm:gap-4'

const SEARCH_INPUT_BASE_CLASSES =
  'w-full pl-10 pr-4 py-2 bg-gray-100 dark:bg-gray-800 border-0 rounded-xl text-sm text-gray-900 dark:text-white placeholder-gray-500'

const SEARCH_INPUT_FOCUS_CLASSES = 'focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all'

const ICON_BUTTON_CLASSES =
  'p-2 sm:p-2.5 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors'

/* eslint-disable max-lines-per-function */
export const Header = () => {
  const { theme, toggleTheme } = useTheme()
  const [searchQuery, setSearchQuery] = useState('')
  const { toggle } = useSidebar()

  return (
    <header className={HEADER_CLASSES}>
      <div className={HEADER_CONTAINER_CLASSES}>
        {/* Mobile menu button */}
        <button
          onClick={toggle}
          className="lg:hidden p-2 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          aria-label="Open menu"
        >
          <Menu className="w-5 h-5 text-gray-600 dark:text-gray-400" />
        </button>

        {/* Search */}
        <div className="flex-1 max-w-md">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
               
              className={`${SEARCH_INPUT_BASE_CLASSES} ${SEARCH_INPUT_FOCUS_CLASSES}`}
            />
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 sm:gap-2">
          {/* Notifications */}
          <button
            className={`relative ${ICON_BUTTON_CLASSES}`}
            aria-label="Notifications"
          >
            <Bell className="w-5 h-5 text-gray-600 dark:text-gray-400" />
            <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full" />
          </button>

          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            className={ICON_BUTTON_CLASSES}
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? (
              <Sun className="w-5 h-5 text-gray-600 dark:text-gray-400" />
            ) : (
              <Moon className="w-5 h-5 text-gray-600 dark:text-gray-400" />
            )}
          </button>
        </div>
      </div>
    </header>
  )
}
