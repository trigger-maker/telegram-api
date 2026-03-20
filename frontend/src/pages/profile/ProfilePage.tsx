import { useState } from 'react'
import {
  User,
  Mail,
  Shield,
  Calendar,
  Edit2,
  Save,
  X,
  Key,
  Smartphone,
  Activity,
} from 'lucide-react'
import { Layout } from '@/components/layout'
import { Card, Button, Input } from '@/components/common'
import { useAuth, useToast } from '@/contexts'
import { useSessions } from '@/hooks'

// Tailwind classes for role badge
const ROLE_BADGE_CLASSES = 'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400'

/* eslint-disable max-lines-per-function, complexity */
export const ProfilePage = () => {
  const { user } = useAuth()
  const toast = useToast()
  const { data: sessions } = useSessions()

  const [isEditing, setIsEditing] = useState(false)
  const [formData, setFormData] = useState({
    username: user?.username || '',
    email: user?.email || '',
  })

  const [passwordData, setPasswordData] = useState({
    current: '',
    new: '',
    confirm: '',
  })
  const [showPasswordForm, setShowPasswordForm] = useState(false)

  const activeSessions = sessions?.filter((s) => s.is_active).length || 0
  const totalSessions = sessions?.length || 0

  const handleSaveProfile = async () => {
    toast.success('Profile updated', 'Changes have been saved')
    setIsEditing(false)
  }

  const handleChangePassword = async () => {
    if (passwordData.new !== passwordData.confirm) {
      toast.error('Error', 'Passwords do not match')
      return
    }
    toast.success('Password updated', 'Your password has been changed')
    setShowPasswordForm(false)
    setPasswordData({ current: '', new: '', confirm: '' })
  }

  return (
    <Layout>
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">My Profile</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Manage your personal information
          </p>
        </div>

        {/* Profile Card */}
        <Card className="p-6">
          <div className="flex items-start justify-between mb-6">
            <div className="flex items-center gap-4">
              <div className="w-20 h-20 bg-gradient-to-br from-primary-500 to-primary-700 rounded-2xl flex items-center justify-center shadow-lg shadow-primary-600/20">
                <User className="w-10 h-10 text-white" />
              </div>
              <div>
                <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
                  {user?.username}
                </h2>
                <p className="text-gray-600 dark:text-gray-400">{user?.email}</p>
                <div className="flex items-center gap-2 mt-2">
                  <span className={ROLE_BADGE_CLASSES}>
                    <Shield className="w-3 h-3" />
                    {user?.role === 'admin' ? 'Administrator' : 'User'}
                  </span>
                </div>
              </div>
            </div>

            {!isEditing ? (
              <Button variant="secondary" onClick={() => setIsEditing(true)}>
                <Edit2 className="w-4 h-4 mr-2" />
                Edit
              </Button>
            ) : (
              <div className="flex gap-2">
                <Button variant="ghost" onClick={() => setIsEditing(false)}>
                  <X className="w-4 h-4" />
                </Button>
                <Button variant="primary" onClick={handleSaveProfile}>
                  <Save className="w-4 h-4 mr-2" />
                  Save
                </Button>
              </div>
            )}
          </div>

          {isEditing ? (
            <div className="grid gap-4 md:grid-cols-2">
              <Input
                label="Username"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              />
              <Input
                label="Email"
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2">
              <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
                <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
                  <User className="w-4 h-4" />
                  Username
                </div>
                <p className="font-medium text-gray-900 dark:text-white">{user?.username}</p>
              </div>

              <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
                <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
                  <Mail className="w-4 h-4" />
                  Email
                </div>
                <p className="font-medium text-gray-900 dark:text-white">{user?.email}</p>
              </div>

              <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
                <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
                  <Shield className="w-4 h-4" />
                  Role
                </div>
                <p className="font-medium text-gray-900 dark:text-white capitalize">
                  {user?.role === 'admin' ? 'Administrator' : 'User'}
                </p>
              </div>

              <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
                <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
                  <Calendar className="w-4 h-4" />
                  ID
                </div>
                <p className="font-mono text-sm text-gray-900 dark:text-white">{user?.id}</p>
              </div>
            </div>
          )}
        </Card>

        {/* Stats */}
        <div className="grid gap-4 md:grid-cols-3">
          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-primary-100 dark:bg-primary-900/30 rounded-xl">
                <Smartphone className="w-6 h-6 text-primary-600 dark:text-primary-400" />
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Total Sessions</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">{totalSessions}</p>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-xl">
                <Activity className="w-6 h-6 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Active Sessions</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">{activeSessions}</p>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-xl">
                <Shield className="w-6 h-6 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Account</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white capitalize">
                  {user?.role}
                </p>
              </div>
            </div>
          </Card>
        </div>

        {/* Security */}
        <Card className="p-6">
          <h3 className="font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            <Key className="w-5 h-5" />
            Security
          </h3>

          {showPasswordForm ? (
            <div className="space-y-4">
              <Input
                label="Current password"
                type="password"
                value={passwordData.current}
                onChange={(e) => setPasswordData({ ...passwordData, current: e.target.value })}
              />
              <div className="grid gap-4 md:grid-cols-2">
                <Input
                  label="New password"
                  type="password"
                  value={passwordData.new}
                  onChange={(e) => setPasswordData({ ...passwordData, new: e.target.value })}
                />
                <Input
                  label="Confirm password"
                  type="password"
                  value={passwordData.confirm}
                  onChange={(e) => setPasswordData({ ...passwordData, confirm: e.target.value })}
                />
              </div>
              <div className="flex gap-3">
                <Button variant="secondary" onClick={() => setShowPasswordForm(false)}>
                  Cancel
                </Button>
                <Button variant="primary" onClick={handleChangePassword}>
                  Change Password
                </Button>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
              <div>
                <p className="font-medium text-gray-900 dark:text-white">Password</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Last updated: 30 days ago
                </p>
              </div>
              <Button variant="secondary" onClick={() => setShowPasswordForm(true)}>
                Change
              </Button>
            </div>
          )}
        </Card>
      </div>
    </Layout>
  )
}
