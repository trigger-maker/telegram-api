/* eslint-disable react/no-unescaped-entities, max-lines-per-function */
import { ExternalLink, Key, Smartphone, Check } from 'lucide-react'
import { Alert, Badge } from '@/components/common'

export const TelegramGuide = () => {
  return (
    <div className="space-y-6">
      <Alert variant="info">
        <div className="space-y-2">
          <p className="font-semibold">What do you need?</p>
          <p className="text-sm">
            To create a Telegram session you need to get your API credentials
            from the official Telegram site.
          </p>
        </div>
      </Alert>

      <div className="space-y-4">
        <div className="flex items-start gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex-shrink-0">
            <span className="text-primary-600 dark:text-primary-400 font-bold">1</span>
          </div>
          <div className="flex-1">
            <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
              Access Telegram API
            </h4>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
              Go to{' '}
              <a
                href="https://my.telegram.org"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary-600 dark:text-primary-400 hover:underline inline-flex items-center gap-1"
              >
                my.telegram.org
                <ExternalLink className="w-3 h-3" />
              </a>
            </p>
            <Badge variant="info">
              <Smartphone className="w-3 h-3 mr-1 inline" />
              You need your Telegram phone number
            </Badge>
          </div>
        </div>

        <div className="flex items-start gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex-shrink-0">
            <span className="text-primary-600 dark:text-primary-400 font-bold">2</span>
          </div>
          <div className="flex-1">
            <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
              Login
            </h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Enter your phone number and the code you will receive by SMS or Telegram.
            </p>
          </div>
        </div>

        <div className="flex items-start gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex-shrink-0">
            <span className="text-primary-600 dark:text-primary-400 font-bold">3</span>
          </div>
          <div className="flex-1">
            <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
              Go to "API development tools"
            </h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              In the main menu, click on "API development tools".
            </p>
          </div>
        </div>

        <div className="flex items-start gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex-shrink-0">
            <span className="text-primary-600 dark:text-primary-400 font-bold">4</span>
          </div>
          <div className="flex-1">
            <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
              Create an application
            </h4>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
              Complete the form with this data:
            </p>
            <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
              <li className="flex items-start gap-2">
                <Check className="w-4 h-4 text-green-600 dark:text-green-400 mt-0.5 flex-shrink-0" />
                <span><strong>App title:</strong> Your App (e.g., "My Telegram Bot")</span>
              </li>
              <li className="flex items-start gap-2">
                <Check className="w-4 h-4 text-green-600 dark:text-green-400 mt-0.5 flex-shrink-0" />
                <span><strong>Short name:</strong> Short name (e.g., "mybot")</span>
              </li>
              <li className="flex items-start gap-2">
                <Check className="w-4 h-4 text-green-600 dark:text-green-400 mt-0.5 flex-shrink-0" />
                <span><strong>Platform:</strong> Select "Other"</span>
              </li>
            </ul>
          </div>
        </div>

        <div className="flex items-start gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-green-100 dark:bg-green-900/30 flex-shrink-0">
            <Check className="w-5 h-5 text-green-600 dark:text-green-400" />
          </div>
          <div className="flex-1">
            <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
              Copy your credentials
            </h4>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
              Once the application is created, you will see:
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                <div className="flex items-center gap-2 mb-1">
                  <Key className="w-4 h-4 text-primary-600 dark:text-primary-400" />
                  <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                    API ID
                  </span>
                </div>
                <code className="text-sm font-mono text-gray-900 dark:text-white">
                  12345678
                </code>
              </div>
              <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                <div className="flex items-center gap-2 mb-1">
                  <Key className="w-4 h-4 text-primary-600 dark:text-primary-400" />
                  <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                    API Hash
                  </span>
                </div>
                <code className="text-sm font-mono text-gray-900 dark:text-white">
                  abc123def456...
                </code>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Alert variant="warning">
        <div className="space-y-2">
          <p className="font-semibold text-sm">⚠️ Important</p>
          <ul className="text-sm space-y-1 ml-4 list-disc">
            <li>Save these credentials securely</li>
            <li>Do not share them with anyone</li>
            <li>You can use the same credentials for multiple sessions</li>
          </ul>
        </div>
      </Alert>
    </div>
  )
}
