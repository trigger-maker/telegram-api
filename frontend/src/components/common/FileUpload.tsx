/* eslint-disable max-lines-per-function */
import { useState, useRef, useCallback } from 'react'
import { Upload, X, Image, Video, Music, FileText, Loader2, Link as LinkIcon } from 'lucide-react'
import { validateFile, FileType } from '@/utils/upload'

interface FileUploadProps {
  type: FileType
  value: string
  onChange: (url: string) => void
  label?: string
  placeholder?: string
  accept?: string
  maxSize?: number
  disabled?: boolean
}

// Tailwind classes for URL input
const URL_INPUT_CLASSES = 'w-full pl-10 pr-4 py-2.5 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all'

// Tailwind classes for toggle buttons
const BUTTON_BASE_CLASSES = 'flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors'
const BUTTON_ACTIVE_CLASSES = 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
const BUTTON_INACTIVE_CLASSES = 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'

const typeConfig = {
  image: {
    icon: Image,
    accept: 'image/jpeg,image/png,image/gif,image/webp',
    label: 'image',
    preview: true,
  },
  video: {
    icon: Video,
    accept: 'video/mp4,video/webm,video/quicktime',
    label: 'video',
    preview: true,
  },
  audio: {
    icon: Music,
    accept: 'audio/mpeg,audio/ogg,audio/wav,audio/mp3',
    label: 'audio',
    preview: false,
  },
  file: {
    icon: FileText,
    accept: 'application/pdf,.doc,.docx,.txt',
    label: 'file',
    preview: false,
  },
}

// Helper component for toggle buttons
const ToggleButtons = ({ useUrl, setUseUrl }: { useUrl: boolean; setUseUrl: (value: boolean) => void }) => (
  <div className="flex gap-2 p-1 bg-gray-100 dark:bg-gray-800 rounded-lg">
    <button
      type="button"
      onClick={() => setUseUrl(true)}
      className={`${BUTTON_BASE_CLASSES} ${useUrl ? BUTTON_ACTIVE_CLASSES : BUTTON_INACTIVE_CLASSES}`}
    >
      <LinkIcon className="w-4 h-4" />
      URL
    </button>
    <button
      type="button"
      onClick={() => setUseUrl(false)}
      className={`${BUTTON_BASE_CLASSES} ${!useUrl ? BUTTON_ACTIVE_CLASSES : BUTTON_INACTIVE_CLASSES}`}
    >
      <Upload className="w-4 h-4" />
      Upload
    </button>
  </div>
)

// Helper component for URL input
const UrlInput = ({
  urlInput,
  handleUrlChange,
  placeholder,
  disabled,
}: {
  urlInput: string
  handleUrlChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  placeholder?: string
  disabled?: boolean
}) => (
  <div className="relative">
    <LinkIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
    <input
      type="url"
      value={urlInput}
      onChange={handleUrlChange}
      placeholder={placeholder}
      disabled={disabled}
      className={URL_INPUT_CLASSES}
    />
  </div>
)

// Helper component for file upload area
const FileUploadArea = ({
  isDragging,
  isLoading,
  Icon,
  config,
  disabled,
  fileInputRef,
  handleInputChange,
  handleDrop,
  handleDragOver,
  handleDragLeave,
}: {
  isDragging: boolean
  isLoading: boolean
  Icon: React.ElementType
  config: { icon: React.ElementType; accept: string; label: string; preview: boolean }
  disabled?: boolean
  fileInputRef: React.RefObject<HTMLInputElement | null>
  handleInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  handleDrop: (e: React.DragEvent) => void
  handleDragOver: (e: React.DragEvent) => void
  handleDragLeave: (e: React.DragEvent) => void
}) => (
  <div
    onDrop={handleDrop}
    onDragOver={handleDragOver}
    onDragLeave={handleDragLeave}
    onClick={() => !disabled && fileInputRef.current?.click()}
    className={`
      relative border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-all
      ${isDragging
        ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
        : 'border-gray-300 dark:border-gray-700 hover:border-primary-400 dark:hover:border-primary-600'
      }
      ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
    `}
  >
    <input
      ref={fileInputRef}
      type="file"
      accept={config.accept}
      onChange={handleInputChange}
      disabled={disabled}
      className="hidden"
    />

    {isLoading ? (
      <div className="flex flex-col items-center">
        <Loader2 className="w-10 h-10 text-primary-600 animate-spin mb-3" />
        <p className="text-sm text-gray-600 dark:text-gray-400">Processing...</p>
      </div>
    ) : (
      <div className="flex flex-col items-center">
        <div className="w-12 h-12 bg-gray-100 dark:bg-gray-800 rounded-xl flex items-center justify-center mb-3">
          <Icon className="w-6 h-6 text-gray-400" />
        </div>
        <p className="text-sm font-medium text-gray-900 dark:text-white mb-1">
          Drag your {config.label} here
        </p>
        <p className="text-xs text-gray-500 dark:text-gray-400">
          or click to select
        </p>
      </div>
    )}
  </div>
)

// Helper component for preview
const Preview = ({
  type,
  value,
  clearFile,
}: {
  type: FileType
  value: string
  clearFile: () => void
}) => (
  <div className="relative rounded-xl overflow-hidden bg-gray-100 dark:bg-gray-800">
    {type === 'image' && (
      <img
        src={value}
        alt="Preview"
        className="w-full h-48 object-cover"
        onError={() => {/* Error handled by parent */}}
      />
    )}
    {type === 'video' && (
      <video
        src={value}
        controls
        className="w-full h-48 object-cover"
        onError={() => {/* Error handled by parent */}}
      />
    )}
    {type === 'audio' && (
      <div className="p-4 flex items-center gap-3">
        <div className="p-3 bg-primary-100 dark:bg-primary-900/30 rounded-lg">
          <Music className="w-6 h-6 text-primary-600 dark:text-primary-400" />
        </div>
        <audio src={value} controls className="flex-1" />
      </div>
    )}
    {type === 'file' && (
      <div className="p-4 flex items-center gap-3">
        <div className="p-3 bg-primary-100 dark:bg-primary-900/30 rounded-lg">
          <FileText className="w-6 h-6 text-primary-600 dark:text-primary-400" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="font-medium text-gray-900 dark:text-white truncate">
            File selected
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
            {value.substring(0, 50)}...
          </p>
        </div>
      </div>
    )}

    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        clearFile()
      }}
      className="absolute top-2 right-2 p-1.5 bg-black/50 hover:bg-black/70 rounded-lg transition-colors"
    >
      <X className="w-4 h-4 text-white" />
    </button>
  </div>
)

export const FileUpload = ({
  type,
  value,
  onChange,
  label,
  placeholder,
  disabled = false,
}: FileUploadProps) => { // eslint-disable max-lines-per-function, complexity
  const [isDragging, setIsDragging] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [useUrl, setUseUrl] = useState(true)
  const [urlInput, setUrlInput] = useState(value)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const config = typeConfig[type]
  const Icon = config.icon

  const handleFileSelect = useCallback(async (file: File) => {
    setError(null)
    setIsLoading(true)

    try {
      const validation = validateFile(file, type)
      if (!validation.valid) {
        setError(validation.error || 'Invalid file')
        return
      }

      const reader = new FileReader()
      reader.onloadend = () => {
        const result = reader.result as string
        onChange(result)
        setIsLoading(false)
      }
      reader.onerror = () => {
        setError('Error reading file')
        setIsLoading(false)
      }
      reader.readAsDataURL(file)
    } catch {
      setError('Error processing file')
      setIsLoading(false)
    }
  }, [type, onChange])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)

    const file = e.dataTransfer.files[0]
    if (file) {
      handleFileSelect(file)
    }
  }, [handleFileSelect])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      handleFileSelect(file)
    }
  }

  const handleUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const url = e.target.value
    setUrlInput(url)
    onChange(url)
  }

  const clearFile = () => {
    onChange('')
    setUrlInput('')
    setError(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const getPlaceholder = (): string => {
    if (placeholder) return placeholder
    const ext = type === 'image' ? 'jpg' : type === 'video' ? 'mp4' : type === 'audio' ? 'mp3' : 'pdf'
    return `https://example.com/${config.label}.${ext}`
  }

  return (
    <div className="space-y-3">
      {label && (
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          {label}
        </label>
      )}

      <ToggleButtons useUrl={useUrl} setUseUrl={setUseUrl} />

      {useUrl ? (
        <UrlInput
          urlInput={urlInput}
          handleUrlChange={handleUrlChange}
          placeholder={getPlaceholder()}
          disabled={disabled}
        />
      ) : (
        <FileUploadArea
          isDragging={isDragging}
          isLoading={isLoading}
          Icon={Icon}
          config={config}
          disabled={disabled}
          fileInputRef={fileInputRef}
          handleInputChange={handleInputChange}
          handleDrop={handleDrop}
          handleDragOver={handleDragOver}
          handleDragLeave={handleDragLeave}
        />
      )}

      {error && (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      )}

      {value && (
        <Preview
          type={type}
          value={value}
          clearFile={clearFile}
        />
      )}
    </div>
  )
}
