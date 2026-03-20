/* eslint-disable max-lines-per-function, complexity */
import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { UserPlus, Mail, Lock, User, Zap, ArrowRight, Check } from 'lucide-react'
import { useAuth } from '@/contexts'
import { Button, Input, Card, Alert } from '@/components/common'
import { ApiException } from '@/types'

export const RegisterPage = () => {
  const navigate = useNavigate()
  const { register, isLoading } = useAuth()

  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [error, setError] = useState('')

  const passwordRequirements = [
    { text: 'At least 8 characters', met: formData.password.length >= 8 },
    { text: 'At least one uppercase letter', met: /[A-Z]/.test(formData.password) },
    { text: 'At least one number', met: /[0-9]/.test(formData.password) },
  ]

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (!formData.username || !formData.email || !formData.password) {
      setError('Please complete all fields')
      return
    }

    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (formData.password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    try {
      await register(formData.username, formData.email, formData.password)
      navigate('/dashboard')
    } catch (err) {
      if (err instanceof ApiException) {
        setError(err.message)
      } else {
        setError('Error creating account. Please try again.')
      }
    }
  }

  return (
    <div className="min-h-screen flex">
      {/* Left side - Form */}
      <div className="flex-1 flex items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
        <div className="w-full max-w-md space-y-8">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-primary-500 to-primary-700 rounded-2xl mb-6 shadow-xl shadow-primary-600/30">
              <Zap className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
              Create account
            </h1>
            <p className="text-gray-600 dark:text-gray-400">
              Register to manage your Telegram sessions
            </p>
          </div>

          <Card className="p-6">
            <form onSubmit={handleSubmit} className="space-y-5">
              {error && (
                <Alert variant="error">
                  {error}
                </Alert>
              )}

              <div className="relative">
                <User className="absolute left-3 top-9 w-5 h-5 text-gray-400" />
                <Input
                  label="Username"
                  type="text"
                  placeholder="your_username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  disabled={isLoading}
                  autoComplete="username"
                  className="pl-10"
                />
              </div>

              <div className="relative">
                <Mail className="absolute left-3 top-9 w-5 h-5 text-gray-400" />
                <Input
                  label="Email"
                  type="email"
                  placeholder="your@email.com"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  disabled={isLoading}
                  autoComplete="email"
                  className="pl-10"
                />
              </div>

              <div className="relative">
                <Lock className="absolute left-3 top-9 w-5 h-5 text-gray-400" />
                <Input
                  label="Password"
                  type="password"
                  placeholder="••••••••"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  disabled={isLoading}
                  autoComplete="new-password"
                  className="pl-10"
                />
              </div>

              {/* Password requirements */}
              {formData.password && (
                <div className="space-y-2">
                  {passwordRequirements.map((req, i) => (
                    <div key={i} className="flex items-center gap-2 text-sm">
                      <div className={`w-4 h-4 rounded-full flex items-center justify-center ${
                        req.met ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'
                      }`}>
                        {req.met && <Check className="w-3 h-3 text-white" />}
                      </div>
                      <span className={req.met ? 'text-green-600 dark:text-green-400' : 'text-gray-500'}>
                        {req.text}
                      </span>
                    </div>
                  ))}
                </div>
              )}

              <div className="relative">
                <Lock className="absolute left-3 top-9 w-5 h-5 text-gray-400" />
                <Input
                  label="Confirm password"
                  type="password"
                  placeholder="••••••••"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                  disabled={isLoading}
                  autoComplete="new-password"
                  className="pl-10"
                />
              </div>

              <Button
                type="submit"
                variant="primary"
                fullWidth
                isLoading={isLoading}
                className="h-12 text-base"
              >
                <UserPlus className="w-5 h-5 mr-2" />
                Create account
              </Button>
            </form>
          </Card>

          <p className="text-center text-sm text-gray-600 dark:text-gray-400">
            Already have an account?{' '}
            <Link to="/login" className="font-semibold text-primary-600 hover:text-primary-500 transition-colors">
              Login
              <ArrowRight className="w-4 h-4 inline ml-1" />
            </Link>
          </p>
        </div>
      </div>

      {/* Right side - Features */}
      <div className="hidden lg:flex lg:flex-1 bg-gradient-to-br from-primary-600 to-primary-800 p-12 items-center justify-center">
        <div className="max-w-md text-white space-y-8">
          <h2 className="text-3xl font-bold">
            Power up your communication with Telegram
          </h2>
          <div className="space-y-6">
            {[
              { title: 'Multi-session', desc: 'Manage multiple Telegram accounts simultaneously' },
              { title: 'Bulk messages', desc: 'Send messages to multiple recipients with delay' },
              { title: 'Webhooks', desc: 'Receive real-time events via webhooks' },
              { title: 'REST API', desc: 'Easy integration with any system' },
            ].map((feature, i) => (
              <div key={i} className="flex items-start gap-4">
                <div className="w-8 h-8 rounded-lg bg-white/20 flex items-center justify-center flex-shrink-0">
                  <Check className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-semibold">{feature.title}</h3>
                  <p className="text-sm text-primary-100">{feature.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
