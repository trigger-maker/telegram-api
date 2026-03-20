/* global File, FileReader, URL */
/* eslint-disable complexity */
import { ALLOWED_FILE_TYPES, MAX_FILE_SIZES } from '@/config/constants'

export type FileType = 'image' | 'video' | 'audio' | 'file'

interface UploadResult {
  success: boolean
  url?: string
  error?: string
  filename?: string
}

/**
 * Validates the file before uploading
 */
export const validateFile = (file: File, type: FileType): { valid: boolean; error?: string } => {
  const allowedTypes = ALLOWED_FILE_TYPES[type]
  const maxSize = MAX_FILE_SIZES[type]

  if (!allowedTypes.includes(file.type)) {
    return {
      valid: false,
      error: `File type not allowed. Allowed types: ${allowedTypes.join(', ')}`,
    }
  }

  if (file.size > maxSize) {
    const maxMB = Math.round(maxSize / (1024 * 1024))
    return {
      valid: false,
      error: `File exceeds maximum allowed size (${maxMB}MB)`,
    }
  }

  return { valid: true }
}

/**
 * Generates a unique name for the file
 */
const generateUniqueFilename = (originalName: string): string => {
  const timestamp = Date.now()
  const random = Math.random().toString(36).substring(2, 8)
  const extension = originalName.split('.').pop()
  const baseName = originalName.replace(/\.[^/.]+$/, '').replace(/[^a-zA-Z0-9]/g, '_')
  return `${baseName}_${timestamp}_${random}.${extension}`
}

/**
 * Converts a file to Base64
 */
export const fileToBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = (error) => reject(error)
  })
}

/**
 * Simulates file upload and returns the public URL
 * In production, this should upload to the server
 */
export const uploadFile = async (file: File, type: FileType): Promise<UploadResult> => {
  try {
    // Validate file
    const validation = validateFile(file, type)
    if (!validation.valid) {
      return { success: false, error: validation.error }
    }

    // Generate unique name
    const filename = generateUniqueFilename(file.name)

    // Determine the folder based on the type
    const folder = type === 'image' ? 'images' : type === 'video' ? 'videos' : type === 'audio' ? 'audio' : 'files'
    
    // In this case, since we don't have an upload endpoint in the backend,
    // we'll use a base64 strategy for small files
    // or the public URL if the file is already online

    // For development: convert to base64 data URL (works for small files)
    // In production you should implement a real upload endpoint
    const base64 = await fileToBase64(file)

    // If the file is too large for base64, return error with instruction
    if (base64.length > 5 * 1024 * 1024) { // 5MB in base64
      return {
        success: false,
        error: 'File too large. Please upload the file to an external server and use the direct URL.',
      }
    }

    // For small files, we can use the data URL
    // In production, this would be the server URL with the corresponding folder
    void folder // Acknowledge folder for future server upload implementation

    return {
      success: true,
      url: base64, // Use base64 as fallback
      filename,
    }
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Error uploading file',
    }
  }
}

/**
 * Formats the file size for display
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * Gets the appropriate icon for the file type
 */
export const getFileTypeIcon = (mimeType: string): string => {
  if (mimeType.startsWith('image/')) return 'Image'
  if (mimeType.startsWith('video/')) return 'Video'
  if (mimeType.startsWith('audio/')) return 'Music'
  if (mimeType === 'application/pdf') return 'FileText'
  return 'File'
}

/**
 * Checks if a URL is valid
 */
export const isValidUrl = (url: string): boolean => {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

/**
 * Extracts the file extension from a URL
 */
export const getFileExtension = (url: string): string => {
  try {
    const pathname = new URL(url).pathname
    const extension = pathname.split('.').pop()
    return extension || ''
  } catch {
    return ''
  }
}
