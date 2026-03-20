import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { API_BASE_URL, AUTH_TOKEN_KEY, REFRESH_TOKEN_KEY } from '@/config/constants'
import { ApiResponse, ApiException } from '@/types'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private handleUnauthorizedError(): void {
    localStorage.removeItem(AUTH_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    window.location.href = '/login'
  }

  private createApiException(error: AxiosError<ApiResponse>): ApiException {
    const { status, data } = error.response!
    return new ApiException(
      data.error?.code || 'UNKNOWN_ERROR',
      data.error?.message || 'Unknown error',
      status,
      data.error?.details
    )
  }

  private handleResponseError(error: AxiosError<ApiResponse>): never {
    if (error.response) {
      const { status } = error.response

      if (status === 401) {
        this.handleUnauthorizedError()
      }

      throw this.createApiException(error)
    }

    throw new ApiException(
      'NETWORK_ERROR',
      'Connection error. Check your internet.',
      0
    )
  }

  private setupInterceptors(): void {
    // Request interceptor - añade el token a cada petición
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem(AUTH_TOKEN_KEY)
        if (token && config.headers) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor - maneja errores globalmente
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError<ApiResponse>) => this.handleResponseError(error)
    )
  }

  public async get<T>(url: string): Promise<T> {
    const response = await this.client.get<ApiResponse<T>>(url)
    return response.data.data as T
  }

  public async post<T>(url: string, data?: unknown): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(url, data)
    return response.data.data as T
  }

  public async put<T>(url: string, data?: unknown): Promise<T> {
    const response = await this.client.put<ApiResponse<T>>(url, data)
    return response.data.data as T
  }

  public async postWithConfig<T>(
    url: string,
    data?: unknown,
    config?: InternalAxiosRequestConfig
  ): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(url, data, config)
    return response.data.data as T
  }

  public async delete<T>(url: string): Promise<T> {
    const response = await this.client.delete<ApiResponse<T>>(url)
    return response.data.data as T
  }
}

export const apiClient = new ApiClient()
