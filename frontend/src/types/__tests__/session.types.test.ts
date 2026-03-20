/**
 * Type definition tests for session types
 * These tests verify type correctness at compile time
 */

import type {
  AuthMethod,
  CreateSessionRequest,
  CreateSessionResponse,
  SubmitPasswordRequest,
  ImportTDataRequest,
  ImportTDataResponse,
  RegenerateQRResponse,
  TelegramSession,
} from '../session.types'

// Type assertion tests - these will fail at compile time if types are incorrect

describe('AuthMethod type', () => {
  it('should accept sms value', () => {
    const method: AuthMethod = 'sms'
    expect(method).toBe('sms')
  })

  it('should accept qr value', () => {
    const method: AuthMethod = 'qr'
    expect(method).toBe('qr')
  })

  it('should accept tdata value', () => {
    const method: AuthMethod = 'tdata'
    expect(method).toBe('tdata')
  })
})

describe('CreateSessionRequest interface', () => {
  it('should accept valid SMS request', () => {
    const request: CreateSessionRequest = {
      phone: '+1234567890',
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      auth_method: 'sms',
    }
    expect(request.auth_method).toBe('sms')
  })

  it('should accept valid QR request', () => {
    const request: CreateSessionRequest = {
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      auth_method: 'qr',
    }
    expect(request.auth_method).toBe('qr')
  })

  it('should accept valid TData request', () => {
    const request: CreateSessionRequest = {
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      auth_method: 'tdata',
    }
    expect(request.auth_method).toBe('tdata')
  })

  it('should allow optional phone', () => {
    const request: CreateSessionRequest = {
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
    }
    expect(request.phone).toBeUndefined()
  })

  it('should allow optional auth_method', () => {
    const request: CreateSessionRequest = {
      phone: '+1234567890',
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
    }
    expect(request.auth_method).toBeUndefined()
  })
})

describe('CreateSessionResponse interface', () => {
  it('should accept response with hint', () => {
    const response: CreateSessionResponse = {
      session: {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'password_required',
        is_active: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      hint: 'My password',
    }
    expect(response.hint).toBe('My password')
  })

  it('should accept response without hint', () => {
    const response: CreateSessionResponse = {
      session: {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'authenticated',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    }
    expect(response.hint).toBeUndefined()
  })

  it('should accept response with phone_code_hash', () => {
    const response: CreateSessionResponse = {
      session: {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'code_sent',
        is_active: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      phone_code_hash: 'test_hash',
    }
    expect(response.phone_code_hash).toBe('test_hash')
  })

  it('should accept response with qr_image_base64', () => {
    const response: CreateSessionResponse = {
      session: {
        id: 'session-1',
        user_id: 'user-1',
        api_id: 12345,
        session_name: 'Test Session',
        auth_state: 'pending',
        is_active: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      qr_image_base64: 'base64string',
    }
    expect(response.qr_image_base64).toBe('base64string')
  })
})

describe('SubmitPasswordRequest interface', () => {
  it('should accept valid password request', () => {
    const request: SubmitPasswordRequest = {
      password: 'test_password',
    }
    expect(request.password).toBe('test_password')
  })
})

describe('ImportTDataRequest interface', () => {
  it('should accept valid request with all fields', () => {
    const files: File[] = [new File([''], 'session.dat')]
    const request: ImportTDataRequest = {
      api_id: 12345,
      api_hash: 'test_hash',
      session_name: 'Test Session',
      tdata: files,
    }
    expect(request.api_id).toBe(12345)
    expect(request.api_hash).toBe('test_hash')
    expect(request.session_name).toBe('Test Session')
    expect(request.tdata).toEqual(files)
  })

  it('should accept request without session_name', () => {
    const files: File[] = [new File([''], 'session.dat')]
    const request: ImportTDataRequest = {
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    }
    expect(request.session_name).toBeUndefined()
  })

  it('should accept multiple files', () => {
    const files: File[] = [
      new File([''], 'session.dat'),
      new File([''], 'key.dat'),
      new File([''], 'auth.dat'),
    ]
    const request: ImportTDataRequest = {
      api_id: 12345,
      api_hash: 'test_hash',
      tdata: files,
    }
    expect(request.tdata.length).toBe(3)
  })
})

describe('ImportTDataResponse interface', () => {
  it('should accept valid response', () => {
    const response: ImportTDataResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        telegram_user_id: 12345,
        username: 'testuser',
        auth_state: 'authenticated',
        auth_method: 'tdata',
      },
    }
    expect(response.session.auth_method).toBe('tdata')
    expect(response.session.is_active).toBe(true)
  })

  it('should accept response without optional fields', () => {
    const response: ImportTDataResponse = {
      session: {
        session_id: 'session-1',
        is_active: true,
        auth_state: 'authenticated',
        auth_method: 'tdata',
      },
    }
    expect(response.session.telegram_user_id).toBeUndefined()
    expect(response.session.username).toBeUndefined()
  })
})

describe('RegenerateQRResponse interface', () => {
  it('should accept valid response', () => {
    const response: RegenerateQRResponse = {
      session_id: 'session-1',
      qr_image_base64: 'new_base64_string',
      message: 'QR code regenerated successfully',
    }
    expect(response.session_id).toBe('session-1')
    expect(response.qr_image_base64).toBe('new_base64_string')
    expect(response.message).toBe('QR code regenerated successfully')
  })
})

describe('TelegramSession interface', () => {
  it('should accept valid session with all fields', () => {
    const session: TelegramSession = {
      id: 'session-1',
      user_id: 'user-1',
      phone_number: '+1234567890',
      api_id: 12345,
      session_name: 'Test Session',
      auth_state: 'authenticated',
      telegram_user_id: 12345,
      telegram_username: 'testuser',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(session.id).toBe('session-1')
    expect(session.phone_number).toBe('+1234567890')
    expect(session.telegram_username).toBe('testuser')
  })

  it('should accept session without optional fields', () => {
    const session: TelegramSession = {
      id: 'session-1',
      user_id: 'user-1',
      api_id: 12345,
      session_name: 'Test Session',
      auth_state: 'authenticated',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(session.phone_number).toBeUndefined()
    expect(session.telegram_user_id).toBeUndefined()
    expect(session.telegram_username).toBeUndefined()
  })
})
